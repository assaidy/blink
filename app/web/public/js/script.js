async function verifyClientID() {
  // TODO: what if clientID was invalid for some reason?
  if (localStorage.getItem("clientID")) return;

  const ua = navigator.userAgent;

  let platform = "Unknown";
  if (ua.includes("Edg")) {
    platform = "Edge";
  } else if (ua.includes("OPR")) {
    platform = "Opera";
  } else if (ua.includes("Firefox")) {
    platform = "Firefox";
  } else if (ua.includes("Chrome") && !ua.includes("Edg")) {
    platform = "Chrome";
  } else if (ua.includes("Safari") && !ua.includes("Chrome")) {
    platform = "Safari";
  }

  let os = "Unknown";
  if (ua.includes("Windows NT")) {
    os = "Windows";
  } else if (ua.includes("Mac OS X")) {
    os = "macOS";
  } else if (ua.includes("Linux")) {
    os = "Linux";
  } else if (ua.includes("Android")) {
    os = "Android";
  } else if (ua.includes("iPhone") || ua.includes("iPad")) {
    os = "iOS";
  }

  try {
    const response = await fetch("/api/v1/auth/clients", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ platform, os }),
    });
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`);
    }
    const data = await response.json();
    localStorage.setItem("clientID", data.clientID);
  } catch (err) {
    console.error(`fetch error:`, err);
  }
}

$listen("DOMContentLoaded", async () => {
  await verifyClientID();

  $listen("htmx:configRequest", (event) => {
    event.detail.headers["X-CSRF-Token"] = $cookies["csrf_token"];
  });
});
