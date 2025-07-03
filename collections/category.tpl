<!DOCTYPE html>
	<html lang="ru">
	<head>
		<meta charset="UTF-8">
		<title>Категории</title>
		<style>
			table { border-collapse: collapse; width: 100%; }
			th, td { border: 1px solid #ccc; padding: 8px; text-align: left; }
			.category { margin-bottom: 20px; }
			h2 { margin-top: 30px; }
			.pagination { margin: 20px 0; text-align: center; }
			.pagination button { margin: 0 2px; padding: 5px 10px; }
		</style>
		<script src="https://code.jquery.com/jquery-3.6.0.min.js"></script>
	</head>
	<body>
		{header}
		<h1>Категории</h1>
		<div id="categoriesContainer"></div>
		<div class="pagination" id="pagination"></div>

		<script>
			const category = "{{CATEGORY}}";
			const url = window.location.protocol + '//' + window.location.host + '/' + category + '.json';
			const ROWS_PER_PAGE = 10;
			let currentPage = 1;
			let data = [];

			function renderTable(page) {
				const container = $('#categoriesContainer');
				container.empty();
				const start = (page - 1) * ROWS_PER_PAGE;
				const end = start + ROWS_PER_PAGE;
				const pageData = data.slice(start, end);

				const categoryDiv = $('<div>').addClass('category');
				const header = $('<h2>').text(category);
				categoryDiv.append(header);

				const table = $('<table>');
				const thead = $('<thead>').html('<tr><th>Заголовок</th><th>Ссылка</th></tr>');
				table.append(thead);
				const tbody = $('<tbody>');
				pageData.forEach(function(item) {
					const row = $('<tr>');
					row.html('<td>' + item.title + '</td><td><a href="' + item.url + '">' + item.url + '</a></td>');
					tbody.append(row);
				});
				table.append(tbody);
				categoryDiv.append(table);
				container.append(categoryDiv);
			}

			function renderPagination() {
				const totalPages = Math.ceil(data.length / ROWS_PER_PAGE);
				const pagination = $('#pagination');
				pagination.empty();
				if (totalPages <= 1) return;
				for (let i = 1; i <= totalPages; i++) {
					const btn = $('<button>').text(i);
					if (i === currentPage) btn.attr('disabled', true);
					btn.on('click', function() {
						currentPage = i;
						renderTable(currentPage);
						renderPagination();
					});
					pagination.append(btn);
				}
			}

			$(document).ready(function() {
				$.getJSON(url, function(json) {
					data = json;
					currentPage = 1;
					renderTable(currentPage);
					renderPagination();
				}).fail(function() {
					$('#categoriesContainer').html('<p>Ошибка загрузки данных категорий</p>');
				});
			});
		</script>
		{footer}
	</body>
	</html>
