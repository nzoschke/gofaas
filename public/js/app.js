var app = new Vue({ // eslint-disable-line
    created: function () {
        this.access_token = this.getCookie("access_token");
    },
    data: {
        access_token: undefined,
        claims: {},
        message: "Goodbye servers!",
        results: "",
    },
    el: "#app",
    methods: {
        getCookie: function (name) {
            var value = "; " + document.cookie;
            var parts = value.split("; " + name + "=");
            if (parts.length === 2) return parts.pop().split(";").shift();
        },
        putWork: function () {
            var vm = this;

            fetch("https://api.gofaas.net/work", {
                method: "POST",
                mode: "cors",
                headers: {
                    "Accept": "application/json",
                    "Authorization": `Bearer ${this.access_token}`,
                    "Content-Type": "application/json"
                },
                body: JSON.stringify({
                    sub: this.claims.sub,
                })
            })
                .then(function(response) {
                    return response.json();
                })
                .then(function (json) {
                    vm.results = json;
                })
                .catch(function (err) {
                    vm.results = err.toString();
                });
        }
    },
    watch: {
        access_token: function (t) {
            var base64Url = t.split(".")[1];
            var base64 = base64Url.replace("-", "+").replace("_", "/");
            this.claims = JSON.parse(window.atob(base64));
        }
    }
});