# Static Website Security 
### With Lambda@Edge and Google OAuth 2.0

Deploying static web content is a solved problem thanks to S3 and CloudFront. However this design has one big side effect: the static content is publicly available to the entire internet. What if our application is more sensitive like an admin app or an internal business tool and we need to restrict access to only people that belong to our company?

Previously we would add a "proxy server" -- an HTTP server like nginx, plus a load balancer and a private network -- to get the access control we need. While this adds security, we lose a lot of simplicity and reliability.

What we need instead is a way to configure access with the proxy we already have: CloudFront. AWS offers a function-as-a-service approach here with Lambda@Edge.

Lambda@Edge is a variant of Lambda that runs "inside" CloudFront. The Lambda function is given an event that describes an HTTP request, and it returns an event that that controls the response to send to the client. This unlocks a lot of options. We can modify the request to the origin and change the body, add headers and/or rewrite the URL. We can modify the response to the users and change the body or add headers. 

This is perfect for authentication. Instead of always responding with our content from S3, we can first check the request for an authorization cookie and redirect to an OAuth provider if missing. We can then handle the OAuth callback and set a cookie for the user. Then subsequent requests will pass the auth check and we can response with content from S3. All with a simple Lambda function.

Let's set this all up for our app...

## AWS Config

We start with the the S3 and CloudFront config from the [static sites](static-sites.md) example, and make a few changes.

First we need to guarantee that our S3 bucket is only accessible through CloudFront. First we make sure to remove any S3 config for public access or a website configuration. Then we add an [S3 Bucket Policy](https://docs.aws.amazon.com/AmazonS3/latest/dev/example-bucket-policies.html) that only grants read access to a [CloudFront Origin Access Identity](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/private-content-restricting-access-to-s3.html) and a CloudFront `S3OriginConfig`. This guarantees that our content is only available via the CloudFront distribution and not over the `s3-website-us-east-1.amazonaws.com` or `s3.amazonaws.com` style URLs.

Next we associate a Lambda function with our distribution. Here we configure a `viewer-request` event to invoke a function when CloudFront receives a request from the viewer. This is the earliest event possible, and gives us the ability to redirect and/or block requests from the origin based on our auth flow.


```yaml
---
AWSTemplateFormatVersion: '2010-09-09'

Resources:
  WebBucket:
    Properties:
      BucketName: !Ref WebDomainName
      WebsiteConfiguration:
        ErrorDocument: 404.html
        IndexDocument: index.html
    Type: AWS::S3::Bucket

  WebBucketPolicyPrivate:
    Properties:
      Bucket: !Ref WebBucket
      PolicyDocument:
        Statement:
          - Action: s3:GetObject
            Effect: Allow
            Principal:
              CanonicalUser: !GetAtt WebOriginAccessIdentity.S3CanonicalUserId
            Resource: !Sub arn:aws:s3:::${WebBucket}/*
            Sid: GetObjectsCloudFront
    Type: AWS::S3::BucketPolicy

  WebDistribution:
    Properties:
      DistributionConfig:
        Aliases:
          - !Ref WebDomainName
        Comment: !Sub Distribution for ${WebBucket}
        DefaultCacheBehavior:
          AllowedMethods:
            - GET
            - HEAD
          Compress: true
          ForwardedValues:
            Cookies:
              Forward: none
            QueryString: true
          LambdaFunctionAssociations:
            EventType: viewer-request
              LambdaFunctionARN: !Ref WebAuthFunction.Version
          TargetOriginId: !Ref WebBucket
          ViewerProtocolPolicy: redirect-to-https
        DefaultRootObject: index.html
        Enabled: true
        HttpVersion: http2
        Origins:
          - DomainName: !Sub ${WebBucket}.s3.amazonaws.com
            Id: !Ref WebBucket
            S3OriginConfig:
              OriginAccessIdentity: !Sub origin-access-identity/cloudfront/${WebOriginAccessIdentity}
        PriceClass: PriceClass_All
        ViewerCertificate:
          AcmCertificateArn: !Ref WebCertificate
          SslSupportMethod: sni-only
    Type: AWS::CloudFront::Distribution

  WebOriginAccessIdentity:
    Condition: WebDomainNameSpecified
    Properties:
      CloudFrontOriginAccessIdentityConfig:
        Comment: !Ref WebBucket
    Type: AWS::CloudFront::CloudFrontOriginAccessIdentity
```
> From [template.yml](template.yml)

## JavaScript Code -- Auth Config

Lambda@Edge has a few caveats, presumably to increase the security and reliability when running "inside" CloudFront. It only supports the `Node.js 6.10` runtime and it doesn't allow environment variables.

For auth there are some obvious config variables: the OAuth client ID and secret, the domain we whitelist (e.g. gofaas.net) and shared secret for signing tokens. We could hard code these into our JS handler, but we can also harness the full power of Lambda@Edge and interact with any other AWS services. So we turn to the [AWS Systems Manager Parameter Store](https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-paramstore.html) to store and access secrets.

Here we define a global variable for our `Params`. On "cold starts" we  ask SSM for the parameters. Note that our Lambda@Edge functions run all around the world, so we need to take care to talk to SSM in the region it was configured. Subsequent requests will "re-enter" the same handler and get the config without hitting SSM.

```js
"use strict";
var AWSXRay = require("aws-xray-sdk");
var AWS = AWSXRay.captureAWS(require("aws-sdk"));

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

var paramsGet = (context) => (new Promise(function (fulfill, reject) {
    // immediate return cached params if defined
    if (Params.AuthDomainName !== undefined) return fulfill();

    var path = "/gofaas/";
    var region = "us-east-1";

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
            // auth logic here
        });
};
```

## JavaScript Code -- Cookies and JWT

Our auth scheme depends on an authorized user presenting a valid JavasScript Web Token (JWT) via a cookie.

Our function is invoked with an event that describes the request, including headers. This makes it easy to read cookies:

```js
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
```

Our function returns a response with a status code and headers. This makes it easy to set a cookie:

```js
var responseCookie = (token, exp, location) => {
    return = {
        status: "302",
        statusDescription: "Found",
        headers: {
            location: [{
                key: "Location",
                value: location,
            }],
            "set-cookie": [{
                key: "Set-Cookie",
                value: `access_token=${token}; expires=${exp.toUTCString()}; path=/`,
            }],
        },
    }
};
```

We can use the [jwt-simple](https://github.com/hokaccha/node-jwt-simple) Node.js library to encode and decode JWT tokens, along with the `AuthHashKey` config from above:

```js
var jwt = require("jwt-simple");

var encode = (profile) => {
    var exp = new Date(new Date().getTime() + 86400000); // 1 day from now
    var key = new Buffer(Params.AuthHashKey, "base64");

    var token = jwt.encode({
        exp: Math.floor(exp / 1000),
        sub: profile.emails[0].value,
    }, key);

    callback(null, responseCookie(token, exp, `https://${host}/`));
};

var decode = (request) => {
    try {
        var key = new Buffer(Params.AuthHashKey, "base64");
        jwt.decode(requestCookie(request, "access_token"), key);
        callback(null, request);
    }
    catch (err) {
        // missing or invalid token
    }
};
```

Now we just have to wire up the auth request and response flow that reads/writes a cookie and encodes/decodes a JWT.

## JavaScript Code -- Google OAuth 2.0

Web authentication is nothing new in the Node.js ecosystem. Here we turn to the [Passport](http://www.passportjs.org/) library which offers "strategies" for many auth providers like Auth0, Google, GitHub and more. For this example we will use the [passport-google-oauth2](https://github.com/jaredhanson/passport-google-oauth2) package which implements Google OAuth 2.0 auth.

First we need to set up our OAuth 2.0 app on Google. Browse to the [Google API Credentials](https://console.developers.google.com/apis/credentials) page to create a project and "OAuth client ID" credentials for a "web application". Then browse to the [Google+ API](https://console.developers.google.com/apis/library/plus.googleapis.com/) library page to enable Google+ as an identity and email scope provider.

<p align="center">
  <img src="img/google-oauth.png" alt="Google OAuth 2.0 App" width="440" />
  <img src="img/google+" alt="Google+ API" width="440" />
</p>

For more information see the [Using OAuth 2.0 for Web Server Applications](https://developers.google.com/identity/protocols/OAuth2WebServer) developer guide.

Next it takes a bit of hacking to use Passport outside of the [Express](https://expressjs.com/) web framework, but thanks to it's flexibility it is possible. We instantiate the `GoogleStrategy` class with our OAuth Client ID and secret, and hook up the strategy callbacks for auth redirect, success or failure to our Lambda response callback.

```js
var GoogleStrategy = require("passport-google-oauth20").Strategy;
var querystring = require("querystring");

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
```