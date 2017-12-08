//= require javascripts/api


(function() {
  $("ul#tokens > li").each(function() {
    var elem = $(this);
    var id = elem.val();

    elem.find("i").click(function() {
      var delConfirm = confirm("Revoke " + elem.text() + "?");
      if (!delConfirm) {
        return false;
      }
      // NOTE delete workaround cause of:
      // https://github.com/yui/yuicompressor/issues/47
      var promise = API.oauth.tokens(id)['delete'];
      promise().then(function() { elem.fadeOut(); });
      return false;
    });
  });

  // handle language switcher events
  $("ul li.language").each(function(i, elem) {
    $(elem).click(function() {
      var cookie = "REVEL_LANG=" + $(this).attr("value");
      document.cookie = cookie;
      window.location.reload();
      return false;
    });
  });
})();
