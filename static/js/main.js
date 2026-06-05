htmx.config.globalViewTransitions = true;

document.body.addEventListener("htmx:responseError", function (event) {
  if (event.detail.xhr.status === 401) {
    window.location.href = "/auth/login";
  }
});
