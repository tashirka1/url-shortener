// htmx config
htmx.config.globalViewTransitions = true;

// clipboard — делегирование на document.body
// ловит клики по .shortLink, включая динамически добавленные через htmx
document.body.addEventListener("click", async (event) => {
  const link = event.target.closest(".shortLink");
  if (!link) return;
  event.preventDefault();
  try {
    await navigator.clipboard.writeText(link.href);
    const old = link.textContent;
    link.textContent = "Copied!";
    setTimeout(() => { link.textContent = old; }, 2000);
  } catch (err) {
    console.error("Failed to copy text: ", err);
  }
});
