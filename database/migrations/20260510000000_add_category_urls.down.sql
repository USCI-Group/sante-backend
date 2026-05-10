-- Remove image_url and banner_url from product_categories
ALTER TABLE "product_categories" DROP COLUMN "image_url";
ALTER TABLE "product_categories" DROP COLUMN "banner_url";
