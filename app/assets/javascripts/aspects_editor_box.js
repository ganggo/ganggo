//= require javascripts/api

(function() {
  API.aspects.get().then(function(aspects) {
    $.each(aspects, function(i, aspect) {
      var li = $('<li><a href="#"></a></li>');
      li.val(aspect.ID);
      li.find("a").append(aspect.Name);
      $(".dropdown-menu").prepend(li);
    });

    $(".dropdown ul li").each(function() {
      var li = $(this);
      li.click(function() {
        $(".dropdown button").html(li.text() + ' <span class="caret"></span>');
      });
    });
  });
})();
