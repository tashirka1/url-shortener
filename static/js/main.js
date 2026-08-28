// htmx config
htmx.config.globalViewTransitions = true;

// clipboard — делегирование на document.body
document.body.addEventListener("click", async (event) => {
  const shortLink = event.target.closest(".shortLink");
  if (shortLink) {
    event.preventDefault();
    try {
      await navigator.clipboard.writeText(shortLink.href);
      if (shortLink._copyTimer) clearTimeout(shortLink._copyTimer);
      if (!shortLink.dataset.origText) shortLink.dataset.origText = shortLink.textContent;
      shortLink.textContent = "Скопировано!";
      shortLink._copyTimer = setTimeout(() => {
        shortLink.textContent = shortLink.dataset.origText;
        delete shortLink.dataset.origText;
        delete shortLink._copyTimer;
      }, 2000);
    } catch (err) {
      console.error("Failed to copy text: ", err);
    }
    return;
  }

  const btn = event.target.closest(".copy-token");
  if (btn) {
    const input = document.getElementById("token-value");
    if (!input) return;
    try {
      await navigator.clipboard.writeText(input.value);
      if (btn._copyTimer) clearTimeout(btn._copyTimer);
      if (!btn.dataset.origText) btn.dataset.origText = btn.textContent;
      btn.textContent = "Скопировано!";
      btn._copyTimer = setTimeout(() => {
        btn.textContent = btn.dataset.origText;
        delete btn.dataset.origText;
        delete btn._copyTimer;
      }, 2000);
    } catch (err) {
      console.error("Failed to copy token: ", err);
    }
  }
});
