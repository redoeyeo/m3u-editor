# Запускаем тесты, сохраняем вывод
go test -v ./domain > OUTPUT 2>&1
TEST_EXIT=$?

if [ $TEST_EXIT -ne 0 ]; then
  # Добавляем разделитель и все Go-файлы с путями и кодом
  echo "" >> OUTPUT
  echo "========================================" >> OUTPUT
  echo "ИСХОДНЫЕ ФАЙЛЫ" >> OUTPUT
  echo "========================================" >> OUTPUT
  echo "" >> OUTPUT

  find . -name "*.go" -not -path "./vendor/*" | sort | while read -r f; do
    echo "--- ФАЙЛ: $f ---" >> OUTPUT
    echo "" >> OUTPUT
    cat "$f" >> OUTPUT
    echo "" >> OUTPUT
    echo "" >> OUTPUT
  done

  # Копируем в буфер обмена (clip для Windows)
  clip < OUTPUT
  echo "Тесты упали. Файл OUTPUT создан и скопирован в буфер обмена."
else
  echo "Все тесты прошли. OUTPUT не нужен."
  rm -f OUTPUT
fi
