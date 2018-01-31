$(function () {
  var elem = $('[name="location"]');
  navigator.geolocation.getCurrentPosition(function(position) {
    elem.val(position.coords.latitude + "," + position.coords.longitude);
  }, function(error) { elem.val("Berlin") });

  $('[data-toggle="datetimepicker"]').datetimepicker({
    minDate: new Date(),
    minTime: new Date(),
    validateOnBlur: false,
    //closeOnDateSelect: true,
    format: 'Y-m-d\\TH:m:sO',
    onShow: function(ct) {
      this.setOptions({timepicker: !$('[name="all_day"]').is(":checked")});
    }
  });

  $('[name="all_day"]').click(function() {
    if ($('[name="all_day"]').is(':checked')) {
      $('[name="end"]').prop('disabled', true);
    } else {
      $('[name="end"]').prop('disabled', false);
    }
  });

  $('[data-toggle="calendar"] > .row > .day > .events > .event').each(function() {
    var popover = $(this).popover({
      container: 'body',
      content: $(this).find('.popover-content').html(),
      html: true
    });
  });

  $('[data-toggle="calendar"] > .row > .day > .events > .event').on('shown.bs.popover', function() {
    var elem = $(this);
    $('.popover:last-child').find('.close').on('click', function() {
      elem.popover('hide');
    });
  });
});
