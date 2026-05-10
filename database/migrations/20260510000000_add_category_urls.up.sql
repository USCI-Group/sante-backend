-- Add image_url and banner_url to product_categories
ALTER TABLE "product_categories" ADD COLUMN "image_url" character varying(500) NULL;
ALTER TABLE "product_categories" ADD COLUMN "banner_url" character varying(500) NULL;
