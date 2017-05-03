//= require javascripts/api

(function() {
  API.aspects.get().then(function(aspects) {
    $.each(aspects, function(i, aspect) {
      var li = $('<li><a href="#"></a></li>');
      li.find("a").html(aspect.Name);
      li.click(function() {
        $(".dropdown button").html(li.text() + ' <span class="caret"></span>');
        $("#aspectID").val(aspect.ID);
      });
      $(".dropdown .dropdown-menu").prepend(li);
    });
  });
})();
