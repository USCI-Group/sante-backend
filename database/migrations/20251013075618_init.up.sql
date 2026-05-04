-- modify "customer_delivery_addresses" table
ALTER TABLE "customer_delivery_addresses" ADD COLUMN "latitude" numeric(12,8) NULL, ADD COLUMN "longitude" numeric(12,8) NULL;
-- modify "outlets" table
ALTER TABLE "outlets" ADD COLUMN "latitude" numeric(12,8) NULL, ADD COLUMN "longitude" numeric(12,8) NULL;
