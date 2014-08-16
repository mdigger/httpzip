# zipcompiler

Простой упаковщик в ZIP-файл.

Отличия только в том, что при компрессии CSS, JS и HTML-файлов они автоматически
минимизируются. Ну и, кроме того, позволяет легко указать mimetype-архива. В
этом случае самым первым файлом в архив будет добавлен файл с именем mimetype и
указанным содержимым. Этот файл добавляется без компрессии и может быть легко
найден в заголовке файла и проверен без предварительного открытия и распаковки
архива.

Для просмотра списка параметров запустите приложение с параметром `-help`:

     $ ./zipcompiler -help
     Usage of ./zipcompiler:
    	  -mime="application/x-webarchive+zip": archive mimetype
    	  -mincss=true: minifying css files
    	  -minhtml=true: minifying html files
    	  -minjs=true: minifying javascript files
    	  -out="": output file name
