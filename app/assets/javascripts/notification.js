(function() {
  $(".notify-element").each(function() {
    var id = $(this).data("id");
    $(this).find("a").each(function() {
      $(this).click(function() {
        var elem = $(this);
        API.notifications(id).put().then(function() {
          window.location.href = elem.attr("href");
        });
        return false;
      });
    });
  });
})();
