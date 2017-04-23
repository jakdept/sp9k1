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
  close();

  console.log("opening");

  target.animate({ marginBottom: "660px" }, 200);
  target.addClass("image-card-selected");
  imageOffset = target.offset();
  console.log(imageOffset);
  console.log(String(imageOffset.top + target.height() + 20)+ "px");
  $("#preview-pane-container").css({
    top: imageOffset.top + target.height() + 20 + "px"
    // "left", imageOffset.left + (target.width / 2)
  });
  $("#preview-pane-container").show(100);
}

$(document).ready(function () {

  $(".preview-pane-close-button ").on("click", close);
  $(".image-card").on("click", function () {
    open($(this));
  });

  $("#preview-pane-container").hide(0);

});