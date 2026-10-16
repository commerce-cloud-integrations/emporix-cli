// Runtime fixes for the static export of the docs site, loaded first by every page
// (scripts/site/pages.py). Served below a base path such as /emporix-cli/ on GitHub
// Pages, the export still holds root links that Mintlify's router builds at runtime
// ("/install"): this script prefixes the base to them as links appear, and turns a click on
// a site link into a full page load, since the router's navigation needs data requests
// (?_rsc=) that a static host cannot answer. Those requests, and the polling of the live
// reload that the export keeps (socket.io), are stopped before they reach the network.
(function () {
  "use strict";
  var base = (document.currentScript && document.currentScript.dataset.base) || "";

  function prefixed(url) {
    if (!base || typeof url !== "string" || url.charAt(0) !== "/" || url.charAt(1) === "/") return url;
    if (url === base || url.indexOf(base + "/") === 0 || url.indexOf(base + "?") === 0 || url.indexOf(base + "#") === 0) return url;
    return base + url;
  }

  function fixLinks(root) {
    var links = root && root.querySelectorAll ? root.querySelectorAll('a[href^="/"]') : [];
    for (var i = 0; i < links.length; i++) {
      var href = links[i].getAttribute("href");
      var fixed = prefixed(href);
      if (fixed !== href) links[i].setAttribute("href", fixed);
    }
  }

  if (base) {
    new MutationObserver(function (mutations) {
      for (var i = 0; i < mutations.length; i++) {
        var m = mutations[i];
        if (m.type === "attributes") {
          fixLinks(m.target.parentNode);
          continue;
        }
        for (var j = 0; j < m.addedNodes.length; j++) {
          var node = m.addedNodes[j];
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
    var link = event.target.closest && event.target.closest("a[href]");
    if (!link || (link.target && link.target !== "_self") || link.hasAttribute("download")) return;
    var href = link.getAttribute("href");
    if (href.charAt(0) !== "/" || href.charAt(1) === "/") return;
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

  var fetchOriginal = window.fetch;
  window.fetch = function (input, init) {
    var url = typeof input === "string" ? input : input instanceof URL ? input.href : input && input.url;
    if (url && isRouterData(url)) {
      return Promise.resolve(new Response(null, { status: 404, statusText: "Not Found" }));
    }
    return fetchOriginal.call(this, typeof input === "string" ? prefixed(input) : input, init);
  };

  var open = XMLHttpRequest.prototype.open;
  XMLHttpRequest.prototype.open = function (method, url) {
    if (typeof url === "string" && url.indexOf("/socket.io/") >= 0) {
      this.send = function () {};
      return;
    }
    return open.apply(this, arguments);
  };
  if (window.WebSocket) {
    var Socket = window.WebSocket;
    window.WebSocket = function (url, protocols) {
      if (String(url).indexOf("/socket.io/") >= 0) throw new Error("live reload is off in the static site");
      return protocols === undefined ? new Socket(url) : new Socket(url, protocols);
    };
    window.WebSocket.prototype = Socket.prototype;
    ["CONNECTING", "OPEN", "CLOSING", "CLOSED"].forEach(function (name) { window.WebSocket[name] = Socket[name]; });
  }
})();
