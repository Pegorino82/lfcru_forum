(function () {
  'use strict';

  // id -> EventSource, for cleanup on element replacement
  var sources = {};

  function connect(elt) {
    var url = elt.getAttribute('sse-connect');
    if (!url) return;

    var id = elt.id;
    if (!id) return;

    // Close stale connection for same element id
    if (sources[id]) {
      sources[id].close();
      delete sources[id];
    }

    var eventName = elt.getAttribute('sse-swap') || 'message';

    var es = new EventSource(url, { withCredentials: true });
    sources[id] = es;

    es.addEventListener(eventName, function (evt) {
      elt.insertAdjacentHTML('beforeend', evt.data);
      if (window.htmx) htmx.process(elt);
    });
  }

  function scan(root) {
    (root || document).querySelectorAll('[sse-connect]').forEach(connect);
  }

  document.addEventListener('DOMContentLoaded', function () { scan(); });

  // Re-scan after HTMX swaps — handles outerHTML replacement of #posts-list
  document.addEventListener('htmx:afterSwap', function (evt) {
    scan(evt.detail.elt);
    if (evt.detail.elt && evt.detail.elt.parentElement) {
      scan(evt.detail.elt.parentElement);
    }
  });
})();
