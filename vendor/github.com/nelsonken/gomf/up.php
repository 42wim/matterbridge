<?php
/**
 *
 * PHP UPLOAD DEMO for gomf test
 * USAGE: 
 *      php -S 127.0.0.1:8080 -t ./
 *
 */

print_r($_FILES);
  
if ($_FILES["picture"]["error"] > 0) {
    echo "Return Code: " . $_FILES["picture"]["error"] . "\n";
} else {
    echo "Upload: " . $_FILES["picture"]["name"] . "\n";
    echo "Type: " . $_FILES["picture"]["type"] . "\n";
    echo "Size: " . ($_FILES["picture"]["size"] / 1024) . " Kb\n";
    echo "Temp file: " . $_FILES["picture"]["tmp_name"] . "\n>";

    if (file_exists($_FILES["picture"]["name"]))
      {
      echo $_FILES["picture"]["name"] . " already exists. \n";
      }
    else
      {
      move_uploaded_file($_FILES["picture"]["tmp_name"], $_FILES["picture"]["name"]);
      echo "Stored in: " . $_FILES["picture"]["name"] . "\n";
      }
}


  

