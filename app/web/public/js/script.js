let cookies = {};

function loadCookies() {
  document.cookie.split(";").forEach((pair) => {
    const trimmed = pair.trim();
    const eqIndex = trimmed.indexOf("=");
    if (eqIndex === -1) return;
    const key = decodeURIComponent(trimmed.slice(0, eqIndex));
    const value = decodeURIComponent(trimmed.slice(eqIndex + 1));
    cookies[key] = value;
  });
  return cookies;
}

function registerClient() {
  if (!localStorage.getItem("clientID")) {
    const userAgent = navigator.userAgent;

    let platform = "Unknown";
    if (userAgent.includes("Chrome") && !userAgent.includes("Edg")) {
      platform = "Chrome";
    } else if (userAgent.includes("Firefox")) {
      platform = "Firefox";
    } else if (userAgent.includes("Safari") && !userAgent.includes("Chrome")) {
      platform = "Safari";
    } else if (userAgent.includes("Edg")) {
      platform = "Edge";
    }

    let os = "Unknown";
    if (userAgent.includes("Linux")) {
      os = "Linux";
    } else if (userAgent.includes("Windows")) {
      os = "Windows";
    } else if (userAgent.includes("Mac") || userAgent.includes("macOS")) {
      os = "macOS";
    } else if (userAgent.includes("Android")) {
      os = "Android";
    } else if (userAgent.includes("iPhone") || userAgent.includes("iPad") || userAgent.includes("iPod")) {
      os = "iOS";
    }

    fetch("/api/v1/auth/clients", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ platform, os, app: "Blink Web v1.0.0" }),
    })
      .then((res) => res.json())
      .then((data) => {
        localStorage.setItem("clientID", data.clientID);
      })
      .catch((err) => console.error("Error registering client:", err));
  }
}

document.addEventListener("DOMContentLoaded", () => {
  loadCookies();

  document.addEventListener("htmx:config-request", (event) => {
    const csrfToken = cookies["csrf_token"];
    if (csrfToken) {
      event.detail.headers["X-CSRF-Token"] = csrfToken;
    }
  });
});
