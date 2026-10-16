// Search for the static export of the docs site: Mintlify's search needs its hosted
// service, so the search button and Cmd/Ctrl+K open Pagefind's search over an index built
// at deploy time (pagefind over the export). Loaded by every page (scripts/site/pages.py).
(function () {
  "use strict";
  var base = (document.currentScript && document.currentScript.dataset.base) || "";
  var overlay = null;
  var loading = null;

  function load() {
    if (loading) return loading;
    loading = new Promise(function (resolve, reject) {
      var css = document.createElement("link");
      css.rel = "stylesheet";
      css.href = base + "/pagefind/pagefind-ui.css";
      document.head.appendChild(css);
      var script = document.createElement("script");
      script.src = base + "/pagefind/pagefind-ui.js";
      script.onload = resolve;
      script.onerror = reject;
      document.head.appendChild(script);
    });
    return loading;
  }

  function build() {
    overlay = document.createElement("div");
    overlay.className = "epx-search-overlay";
    overlay.setAttribute("role", "dialog");
    overlay.setAttribute("aria-modal", "true");
    overlay.setAttribute("aria-label", "Search");
    overlay.innerHTML = '<div class="epx-search-panel"><div id="epx-search"></div></div>';
    overlay.addEventListener("mousedown", function (event) {
      if (event.target === overlay) close();
    });
    document.body.appendChild(overlay);
    new window.PagefindUI({
      element: "#epx-search",
      baseUrl: base + "/",
      showSubResults: true,
      showImages: false,
      resetStyles: false,
      autofocus: true,
    });
  }

  function open() {
    load().then(function () {
      if (!overlay) build();
      overlay.classList.add("is-open");
      document.documentElement.classList.add("epx-search-open");
      var input = overlay.querySelector("input");
      if (input) {
        input.focus();
        input.select();
      }
    }, function () {
      console.error("epx search: the Pagefind index is missing; build the site with scripts/site/build.sh");
    });
  }

  function close() {
    if (!overlay) return;
    overlay.classList.remove("is-open");
    document.documentElement.classList.remove("epx-search-open");
  }

  function isOpen() {
    return overlay && overlay.classList.contains("is-open");
  }

  // Capture phase, so that Mintlify's own handlers never see these events.
  document.addEventListener("click", function (event) {
    var button = event.target.closest && event.target.closest("#search-bar-entry, #search-bar-entry-mobile");
    if (!button) return;
    event.preventDefault();
    event.stopImmediatePropagation();
    open();
  }, true);

  window.addEventListener("keydown", function (event) {
    if ((event.metaKey || event.ctrlKey) && (event.key === "k" || event.key === "K")) {
      event.preventDefault();
      event.stopImmediatePropagation();
      if (isOpen()) close(); else open();
    } else if (event.key === "Escape" && isOpen()) {
      event.preventDefault();
      event.stopImmediatePropagation();
      close();
    }
  }, true);
})();
