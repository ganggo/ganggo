(function() {
  var tabs = $(".nav.nav-tabs");
  var widthChildren = 0;
  tabs.children().each(function() {
    widthChildren += $(this).width();
  });

  if (widthChildren > tabs.width()) {
    tabs.siblings("i").each(function() {
      $(this).show();
      $(this).click(function() {
        var curPos = tabs.scrollLeft();
        if ($(this).hasClass("left")) {
          tabs.scrollLeft(curPos - 100);
        } else {
          tabs.scrollLeft(curPos + 100);
        }
        return false;
      });
    });
  }
})();
