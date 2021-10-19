const {csrf, PageIsProjects} = window.config;

export default async function initProject() {
  if (!PageIsProjects) {
    return;
  }

  const {Sortable} = await import(/* webpackChunkName: "sortable" */'sortablejs');
  const boardColumns = document.getElementsByClassName('board-column');

  new Sortable(
    document.getElementsByClassName('board')[0],
    {
      group: 'board-column',
      draggable: '.board-column',
      animation: 150,
      ghostClass: 'card-ghost',
      onSort: () => {
        const board = document.getElementsByClassName('board')[0];
        const boardColumns = board.getElementsByClassName('board-column');

        boardColumns.forEach((column, i) => {
          if (parseInt($(column).data('sorting')) !== i) {
            $.ajax({
              url: $(column).data('url'),
              data: JSON.stringify({sorting: i, color: rgbToHex($(column).css('backgroundColor'))}),
              headers: {
                'X-Csrf-Token': csrf,
                'X-Remote': true,
              },
              contentType: 'application/json',
              method: 'PUT',
            });
          }
        });
      },
    },
  );

  for (const column of boardColumns) {
    new Sortable(
      column.getElementsByClassName('board')[0],
      {
        group: 'shared',
        animation: 150,
        ghostClass: 'card-ghost',
        onAdd: (e) => {
          $.ajax(`${e.to.dataset.url}/${e.item.dataset.issue}`, {
            headers: {
              'X-Csrf-Token': csrf,
              'X-Remote': true,
            },
            contentType: 'application/json',
            type: 'POST',
            error: () => {
              e.from.insertBefore(e.item, e.from.children[e.oldIndex]);
            },
          });
        },
      },
    );
  }

  $('.edit-project-board').each(function () {
    const projectHeader = $(this).closest('.board-column-header');
    const projectTitleLabel = projectHeader.find('.board-label');
    const projectTitleInput = $(this).find(
      '.content > .form > .field > .project-board-title',
    );
    const projectColorInput = $(this).find('.content > .form > .field  #new_board_color');
    const boardColumn = $(this).closest('.board-column');

    if (boardColumn.css('backgroundColor')) {
      setLabelColor(projectHeader, rgbToHex(boardColumn.css('backgroundColor')));
    }

    $(this)
      .find('.content > .form > .actions > .red')
      .on('click', function (e) {
        e.preventDefault();

        $.ajax({
          url: $(this).data('url'),
          data: JSON.stringify({title: projectTitleInput.val(), color: projectColorInput.val()}),
          headers: {
            'X-Csrf-Token': csrf,
            'X-Remote': true,
          },
          contentType: 'application/json',
          method: 'PUT',
        }).done(() => {
          projectTitleLabel.text(projectTitleInput.val());
          projectTitleInput.closest('form').removeClass('dirty');
          if (projectColorInput.val()) {
            setLabelColor(projectHeader, projectColorInput.val());
          }
          boardColumn.attr('style', `background: ${projectColorInput.val()}!important`);
          $('.ui.modal').modal('hide');
        });
      });
  });

  $(document).on('click', '.set-default-project-board', async function (e) {
    e.preventDefault();

    await $.ajax({
      method: 'POST',
      url: $(this).data('url'),
      headers: {
        'X-Csrf-Token': csrf,
        'X-Remote': true,
      },
      contentType: 'application/json',
    });

    window.location.reload();
  });

  $('.delete-project-board').each(function () {
    $(this).click(function (e) {
      e.preventDefault();

      $.ajax({
        url: $(this).data('url'),
        headers: {
          'X-Csrf-Token': csrf,
          'X-Remote': true,
        },
        contentType: 'application/json',
        method: 'DELETE',
      }).done(() => {
        window.location.reload();
      });
    });
  });

  $('#new_board_submit').click(function (e) {
    e.preventDefault();

    const boardTitle = $('#new_board');
    const projectColorInput = $('#new_board_color_picker');

    $.ajax({
      url: $(this).data('url'),
      data: JSON.stringify({title: boardTitle.val(), color: projectColorInput.val()}),
      headers: {
        'X-Csrf-Token': csrf,
        'X-Remote': true,
      },
      contentType: 'application/json',
      method: 'POST',
    }).done(() => {
      boardTitle.closest('form').removeClass('dirty');
      window.location.reload();
    });
  });
}

function setLabelColor(label, color) {
  const red = getRelativeColor(parseInt(color.substr(1, 2), 16));
  const green = getRelativeColor(parseInt(color.substr(3, 2), 16));
  const blue = getRelativeColor(parseInt(color.substr(5, 2), 16));
  const luminance = 0.2126 * red + 0.7152 * green + 0.0722 * blue;

  if (luminance > 0.179) {
    label.removeClass('light-label').addClass('dark-label');
  } else {
    label.removeClass('dark-label').addClass('light-label');
  }
}

/**
 * Inspired by W3C recommandation https://www.w3.org/TR/WCAG20/#relativeluminancedef
 */
function getRelativeColor(color) {
  color /= 255;
  return color <= 0.03928 ? color / 12.92 : ((color + 0.055) / 1.055) ** 2.4;
}

function rgbToHex(rgb) {
  rgb = rgb.match(/^rgb\((\d+),\s*(\d+),\s*(\d+)\)$/);
  return `#${hex(rgb[1])}${hex(rgb[2])}${hex(rgb[3])}`;
}

function hex(x) {
  const hexDigits = ['0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f'];
  return Number.isNaN(x) ? '00' : hexDigits[(x - x % 16) / 16] + hexDigits[x % 16];
}
