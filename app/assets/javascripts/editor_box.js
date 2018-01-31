//= require javascripts/api

(function() {
  var form = $('#post-editor-form');
  if (form !== undefined) {
    form.ajaxForm(function(result) {
      window.location = "/posts/" + result['Guid'];
    });
    loadAspectList();
  }

  var form = $('#comment-editor-form');
  if (form !== undefined) {
    form.ajaxForm(function(result) {
      location.reload();
    });
  }

  var form = $('#calendar-editor-form');
  if (form !== undefined) {
    form.ajaxForm(function(result) {
      location.reload();
    });
    loadAspectList();
  }

  function loadAspectList() {
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
})();
