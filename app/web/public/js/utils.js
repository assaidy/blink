const $listen = document.addEventListener;
const $select = document.querySelector;

const $cookies = {};
document.cookie.split("; ").forEach((cookie) => {
  const separatorIndex = cookie.indexOf("=");
  const name = decodeURIComponent(cookie.slice(0, separatorIndex));
  const value = decodeURIComponent(cookie.slice(separatorIndex + 1));
  $cookies[name] = value;
});
