"use strict";
exports.__esModule = true;
exports.Log = void 0;
function Log(severity, message, baseUrl) {
    if (baseUrl === void 0) { baseUrl = ""; }
    fetch(baseUrl + "/goscope2/js", {
        method: "post",
        headers: { Token: "104365" },
        body: JSON.stringify({
            severity: "WARNING",
            message: "This is a test from javascript"
        })
    });
}
exports.Log = Log;
