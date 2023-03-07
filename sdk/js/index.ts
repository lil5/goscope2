export type Severity = "INFO" | "WARNING" | "ERROR" | "FATAL";

export function Log(severity: Severity, message: string, baseUrl: string = "") {
  fetch(baseUrl + "/goscope2/js", {
    method: "post",
    headers: { Token: "104365" },
    body: JSON.stringify({
      severity: "WARNING",
      message: "This is a test from javascript",
    }),
  });
}
