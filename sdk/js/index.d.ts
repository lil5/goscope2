export type Severity = "INFO" | "WARNING" | "ERROR" | "FATAL";

declare global {
  interface Window {
    goscope2: {
      token: string;
      baseUrl: string;
      New(token: string, baseUrl: string): void;
      Log(severity: Severity, message: string): void;
    };
  }
}
