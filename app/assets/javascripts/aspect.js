//= require javascripts/api

(function() {
  var personId = $("#personID").val();
  API.people(personId).aspects.get().then(function(personInAspects) {
    API.aspects.get().then(function(aspects) {
      $.each(aspects, function(i, aspect) {
        var li = $('<li><a href="#"><span style="display:none" class="glyphicon glyphicon-ok" aria-hidden="true"></span> </a></li>');
        li.click(function() {
          API.people(personId).aspects(aspect.ID).post().then(function() {
            li.addClass("active");
            li.find("span").toggle();
          });
        });

        for (var i in personInAspects) {
          if (personInAspects[i].ID == aspect.ID) {
            // person is already in an aspect
            li.addClass("active");
            li.find("span").toggle();
            // delete instead of create call
            li.off('click');
            li.click(function() {
              // NOTE delete workaround cause of:
              // https://github.com/yui/yuicompressor/issues/47
              promise = API.people(personId).aspects(aspect.ID)['delete']
              promise().then(function() {
                li.removeClass("active");
                li.find("span").toggle();
              });
            });
          }
        }

        li.find("a").append(aspect.Name);
        $(".dropdown-menu").prepend(li);
      });
    });
  });

  $(".toggle-create-aspect").click(function() {
    $(".create-aspect").toggle();
  });

  $("#create-aspect").submit(function() {
    API.aspects.post({
      name: $("#aspectName").val(),
      personId: $("#personID").val()
    }).then(function() {
      window.location.reload();
    });
    return false;
  });
})();
