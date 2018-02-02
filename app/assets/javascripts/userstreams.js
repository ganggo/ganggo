(function() {
  // on confirm trigger reshare api call
  $("a#user-stream").popover({
    html: true,
    title: $("a#user-stream .popover-head").html(),
    content: $("a#user-stream .popover-content").html(),
    placement: "bottom",
    container: "body"
  });

  var form = $("#user-stream-create");
  if (form !== undefined) {
    form.ajaxForm(function(result) {
      window.location.reload();
    });
  }

  var form = $("#user-stream-delete");
  if (form !== undefined) {
    form.find("select").on("change", function() {
      form.find("button").prop("disabled", false);
      form.ajaxForm({
        type: "DELETE",
        url: form.attr("action") + form.find("select").val(),
        success: function() {
          window.location.reload();
        }
      });
    });
  }
})();
