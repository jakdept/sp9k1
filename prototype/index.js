function close() {
  console.log("closing");
  if ($(".image-card-selected").length) {
    $("#preview-pane-container").hide(100, function () {
      $(".image-card-selected").animate({ marginBottom: "10px" }, 200);
      $(".image-card-selected").removeClass("image-card-selected");
    });
  }
}

function open(target) {
  console.log("opening")

  newTop = target.offset().top + target.height() + 20 + "px";
  $("#preview-pane-container").css({ top: newTop });
  target.addClass("image-card-selected");

  target.animate({ marginBottom: "660px" }, 200);
  setTimeout(function () {
    $("#preview-pane-container").show(100);
  }, 100);
}

function imageClick(target) {
  close();
  setTimeout(function() {
    open(target);
  }, 200)
}

$(document).ready(function () {

  $(".preview-pane-close-button ").on("click", close);
  $(".image-card").on("click", function () {
    imageClick($(this));
  });

  $("#preview-pane-container").hide(0);

});