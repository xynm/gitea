export function initRepoBranchButton() {
  $('.show-create-branch-modal.button').on('click', function () {
    $('#create-branch-form')[0].action = $('#create-branch-form').data('base-action') + $(this).data('branch-from');
    $('#modal-create-branch-from-span').text($(this).data('branch-from'));
    $($(this).data('modal')).modal('show');
  });
}
