(function() {
  $(".notify-element").click(function() {
    API.notifications($(this).data("id")).put();
  });
})();
