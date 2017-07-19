//= require javascripts/api

(function() {
  API.aspects.get().then(function(aspects) {
    $.each(aspects, function(i, aspect) {
      $('#aspect-list').append(
        $('<option>', {
          value: aspect.ID,
          text : aspect.Name
        })
      );
    });
  });
})();
