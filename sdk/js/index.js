window.goscope2 = {
  token: "",
  baseUrl: "",
  New(token, baseUrl = "") {
    this.token = token;
    this.baseUrl = baseUrl;
  },
  Log(severity, message) {
    fetch(this.baseUrl + "/goscope2/js", {
      method: "post",
      body: JSON.stringify({ severity, message, token: this.token }),
    });
  },
};
