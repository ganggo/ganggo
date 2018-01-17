(function() {
  $(".notify-element").each(function() {
    var id = $(this).data("id");
    $(this).find("a").click(function() {
      var elem = $(this);
      API.notifications(id).put().then(function() {
        window.location.href = elem.attr("href");
      });
      return false;
    });
    $(this).click(function() {
      var elem = $(this);
      API.notifications(id).put().then(function() {
        elem.remove();
      });
      return false;
    });
  });
})();
