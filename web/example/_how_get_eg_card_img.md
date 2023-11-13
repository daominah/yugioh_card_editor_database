# Example card image from Master Duel

Steps to get example card with resolution 1180x1720 from Master Duel.

### 1. Screenshot cards from Master Duel

* Change Master Duel Settings `Resolution` to `3840x2160` and `Quality` to
  `High Resolition`.
* From Deck editor, click zoom on cards you want, take screenshots, save them
  in a same directory for processing later.

### 2. Process screenshots

Screenshots contain card image at point (left 264, top 400) with size 894x1303,
with the exception Royal cards locate at (left 267, top 400).

Use `imagemagick-6` commands to process screenshots:

````bash
# batch processing from screenshots dir:

export cropDir=crop_894x1303+264+400
export resizeDir=resize_1180x1720
export finalDir=final_jpg_1180x1720

for input in *.png; do echo "$input"; done
for input in *.png; do convert "$input" -crop 894x1303+264+400 "${cropDir}/${input}_cropped.png"; done

for input in $(ls ${cropDir}); do echo "$input"; done
for input in $(ls ${cropDir}); do convert "${cropDir}/$input" -resize 1180 "${resizeDir}/${input}_1180.png"; done

mkdir -p $finalDir
cp ${resizeDir}/* $finalDir/
cd $finalDir
mogrify -quality 96 -format jpg *.png
rm ./*.png
for file in *; do mv "${file}" "${file/.png_cropped.png_1180/}"; done
````

````bash
# process 1 screenshot
export input=eg_monster_effect_long_text.png

export cropDir=crop_894x1303+264+400
export resizeDir=resize_1180x1720
export finalDir=final_jpg_1180x1720

convert "$input" -crop 894x1303+267+400 "${cropDir}/${input}_cropped.png"
convert "${cropDir}/${input}_cropped.png" -resize 1180 "${resizeDir}/${input}_1180.png"
cp "${resizeDir}/${input}_1180.png" "${finalDir}/${input}_1180.png"
convert "${finalDir}/${input}_1180.png" -quality 96 "${finalDir}/${input}.jpg"
rm "${finalDir}/${input}_1180.png"
````
