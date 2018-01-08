//= require javascripts/api

(function() {
  var form = $('#post-editor-form');
  if (form !== undefined) {
    form.ajaxForm(function(result) {
      window.location = "/posts/" + result['Guid'];
    });

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
  }

  var form = $('#comment-editor-form');
  if (form !== undefined) {
    form.ajaxForm(function(result) {
      location.reload();
    });
  }
})();
