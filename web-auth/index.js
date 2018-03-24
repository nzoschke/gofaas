"use strict";
var AWSXRay = require("aws-xray-sdk");
var AWS = AWSXRay.captureAWS(require("aws-sdk"));
var GoogleStrategy = require("passport-google-oauth20").Strategy;
var jwt = require("jwt-simple");
var querystring = require("querystring");

AWSXRay.captureHTTPsGlobal(require("http"));

// global var reused across invocations
var Params = {
    AuthDomainName: undefined,
    AuthHashKey: undefined,
    OAuthClientId: undefined,
    OAuthClientSecret: undefined,
    Scope: [
        "https://www.googleapis.com/auth/plus.login",
        "https://www.googleapis.com/auth/userinfo.email",
    ],
};

var requestCookie = (request, name) => {
    var headers = request.headers;

    var cookies = {};
    if (headers.cookie) {
        headers.cookie[0].value.split(";").forEach((cookie) => {
            if (cookie) {
                const parts = cookie.split("=");
                cookies[parts[0].trim()] = parts[1].trim();
            }
        });
    }

    return cookies[name];
};

var responseCookie = (token, exp, location) => {
    var r = responseRedirect(location);
    r.headers["set-cookie"] = [{
        key: "Set-Cookie",
        value: `access_token=${token}; expires=${exp.toUTCString()}; path=/`,
    }];
    return r;
};

var responseError = (err) => ({
    status: "401",
    statusDescription: "Unauthorized",
    headers: {
        "content-type": [{
            key: "Content-Type",
            value: "text/html",
        }],
    },
    body: JSON.stringify(err),
});

var responseRedirect = (location) => ({
    status: "302",
    statusDescription: "Found",
    headers: {
        location: [{
            key: "Location",
            value: location,
        }],
    },
});

var auth = (request, callback) => {
    var host = request.headers.host[0].value;
    var query = querystring.parse(request.querystring);

    var opts = {
        clientID: Params.OAuthClientId,
        clientSecret: Params.OAuthClientSecret,
        callbackURL: `https://${host}/auth`,
    };

    var s = new GoogleStrategy(opts, (token, tokenSecret, profile, done) => {
        profile.emails.forEach((email) => {
            if (email.value.endsWith(Params.AuthDomainName)) {
                return done(null, profile); // call success with profile
            }
        });

        // call fail with warning
        done(null, false, {
            name: "UserError",
            message: "Email is not a member of the domain",
            status: "401",
        });
    });

    s.error = (err) => {
        callback(null, responseError(err));
    };

    s.fail = (warning) => {
        callback(null, responseError(warning));
    };

    s.redirect = (url) => {
        callback(null, responseRedirect(url));
    };

    s.success = (profile) => {
        var exp = new Date(new Date().getTime() + 86400000); // 1 day from now
        var key = new Buffer(Params.AuthHashKey, "base64");

        var token = jwt.encode({
            exp: Math.floor(exp / 1000),
            sub: profile.emails[0].value,
        }, key);

        callback(null, responseCookie(token, exp, `https://${host}/`));
    };

    s.authenticate({ query }, { scope: Params.Scope });
};

var paramsGet = (context) => (new Promise(function (fulfill, reject) {
    // immediate return cached params if defined
    if (Params.AuthDomainName !== undefined) return fulfill();

    // infer the stack and region from the functionName, e.g. "us-east-1.gofaas-WebAuthFunction"
    var path = "/gofaas/";
    var region = "us-east-1";
    var parts = context.functionName.split(".");
    if (parts.length == 2) {
        path = `/${parts[1].replace("-WebAuthFunction", "")}/`;
        region = parts[0];
    }

    new AWS.SSM({ region })
        .getParametersByPath({ Path: path })
        .promise()
        .then(data => fulfill(
            data.Parameters.forEach((p) => {
                Params[p.Name.slice(path.length)] = p.Value;
            })
        ))
        .catch(err => (reject(err)));
}));

exports.handler = (event, context, callback) => {
    paramsGet(context)
        .then(() => {
            var request = event.Records[0].cf.request;
            var host = request.headers.host[0].value;

            // explicitly call middleware
            if (request.uri === "/auth")
                return auth(request, callback);

            // explicitly expire token
            if (request.uri === "/auth/expire")
                return callback(null, responseCookie("", new Date(0), `https://${host}/auth`));

            // if token is valid make original request
            // if invalid call middleware
            try {
                var key = new Buffer(Params.AuthHashKey, "base64");
                jwt.decode(requestCookie(request, "access_token"), key);
                callback(null, request);
            }
            catch (err) {
                auth(request, callback);
            }
        })
        .catch(err => (callback(err)));
};