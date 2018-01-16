//= require javascripts/api

(function() {
  var personID = $("[name='personID']").val();
  API.people(personID).aspects.get().then(function(aspects) {
    for (var i = 0; i < aspects.length; i++) {
      var name = aspects[i].Name;
      var id = aspects[i].ID;
      var li = $("<li>").attr("data-id", id).text(name);
      $("#aspectList").append(li);
    }

    $("#aspect-form").submit(function() {
      var aspectID = -1;
      var aspectName = $("[name='aspectName']").val();
      for (var i = 0; i < aspects.length; i++) {
        var name = aspects[i].Name;
        if (aspectName.toUpperCase() === name.toUpperCase()) {
          aspectID = aspects[i].ID;
          break;
        }
      }

      if (aspectID > -1) {
        // NOTE delete workaround cause of:
        // https://github.com/yui/yuicompressor/issues/47
        promise = API.people(personID).aspects(aspectID)['delete'];
        promise().then(function(m) {
          $("[data-id='"+aspectID+"']").remove();
        });
      } else {
        API.aspects.get().then(function(aspects) {
          var id = -1;
          var aspectName = $("[name='aspectName']").val();
          for (var i = 0; i < aspects.length; i++) {
            var name = aspects[i].Name;
            if (aspectName.toUpperCase() === name.toUpperCase()) {
              id = aspects[i].ID;
              break;
            }
          }
          var addMembership = function(mid) {
            API.people(personID).aspects(mid).post()
              .then(function(membership) {
                var name = $("[name='aspectName']").val();
                var id = membership.AspectID;
                var li = $("<li>").attr("data-id", id).text(name);
                $("#aspectList").append(li);
            });
          }

          if (id == -1) {
            // create aspect name
            API.aspects.post({aspect_name: aspectName}).then(function(aspect) {
              addMembership(aspect.ID);
            });
          } else {
            addMembership(id);
          }
        });
      }
      return false;
    });
  });

  API.aspects.get().then(function(aspects) {
    var tags = [];
    for (var i = 0; i < aspects.length; i++) {
      tags.push(aspects[i].Name);
    }
    // XXX requires jquery-ui
    //$("[name='aspectName']").autocomplete({source: tags});
  });
})();
