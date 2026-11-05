// Runtime fixes for the static export of the docs site, loaded first by every page
// (.github/scripts/site/pages.py). Served below a base path such as /emporix-cli/ on GitHub
// Pages, the export still holds root links that Mintlify's router builds at runtime
// ("/install"): this script prefixes the base to them as links appear, and turns a click on
// a site link into a full page load, since the router's navigation needs data requests
// (?_rsc=) that a static host cannot answer. Those requests, and the polling of the live
// reload that the export keeps (socket.io), are stopped before they reach the network.
(function () {
  "use strict";
  var base = document.currentScript?.dataset.base || "";

  function prefixed(url) {
    if (!base || typeof url !== "string" || !url.startsWith("/") || url.startsWith("//")) return url;
    if (url === base || url.startsWith(base + "/") || url.startsWith(base + "?") || url.startsWith(base + "#")) return url;
    return base + url;
  }

  function fixLinks(root) {
    var links = root?.querySelectorAll ? root.querySelectorAll('a[href^="/"]') : [];
    for (var link of links) {
      var href = link.getAttribute("href");
      var fixed = prefixed(href);
      if (fixed !== href) link.setAttribute("href", fixed);
    }
  }

  if (base) {
    new MutationObserver(function (mutations) {
      for (var m of mutations) {
        if (m.type === "attributes") {
          fixLinks(m.target.parentNode);
          continue;
        }
        for (var node of m.addedNodes) {
          if (node.nodeType !== 1) continue;
          if (node.matches('a[href^="/"]')) fixLinks(node.parentNode);
          fixLinks(node);
        }
      }
    }).observe(document.documentElement, { childList: true, subtree: true, attributes: true, attributeFilter: ["href"] });
  }

  // Capture phase: the router never sees a click on a site link.
  document.addEventListener("click", function (event) {
    if (event.defaultPrevented || event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return;
    var link = event.target.closest?.("a[href]");
    if (!link || (link.target && link.target !== "_self") || link.hasAttribute("download")) return;
    var href = link.getAttribute("href");
    if (!href.startsWith("/") || href.startsWith("//")) return;
    event.preventDefault();
    event.stopImmediatePropagation();
    var target = prefixed(href);
    var hash = target.indexOf("#");
    var path = hash < 0 ? target : target.slice(0, hash);
    var here = location.pathname.replace(/\/$/, "");
    if (hash >= 0 && (path === "" || path.replace(/\/$/, "") === here)) {
      location.hash = target.slice(hash);
    } else {
      location.assign(target);
    }
  }, true);

  function isRouterData(url) {
    // Router data (_rsc) and Mintlify's own API (the signed-in user), which only the hosted site answers.
    return /[?&]_rsc=/.test(url) || /\/_mintlify\/api\//.test(url);
  }

  function urlOf(input) {
    if (typeof input === "string") return input;
    if (input instanceof URL) return input.href;
    return input?.url;
  }

  var fetchOriginal = window.fetch;
  window.fetch = function (input, init) {
    var url = urlOf(input);
    if (url && isRouterData(url)) {
      return Promise.resolve(new Response(null, { status: 404, statusText: "Not Found" }));
    }
    return fetchOriginal.call(this, typeof input === "string" ? prefixed(input) : input, init);
  };

  var open = XMLHttpRequest.prototype.open;
  XMLHttpRequest.prototype.open = function (method, url) {
    if (typeof url === "string" && url.includes("/socket.io/")) {
      this.send = function () {};
      return;
    }
    return open.apply(this, arguments);
  };
  if (window.WebSocket) {
    var Socket = window.WebSocket;
    window.WebSocket = function (url, protocols) {
      if (String(url).includes("/socket.io/")) throw new Error("live reload is off in the static site");
      return protocols === undefined ? new Socket(url) : new Socket(url, protocols);
    };
    window.WebSocket.prototype = Socket.prototype;
    ["CONNECTING", "OPEN", "CLOSING", "CLOSED"].forEach(function (name) { window.WebSocket[name] = Socket[name]; });
  }
})();
