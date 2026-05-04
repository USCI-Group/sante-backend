-- reverse: modify "outlets" table
ALTER TABLE "outlets" DROP COLUMN "longitude", DROP COLUMN "latitude";
-- reverse: modify "customer_delivery_addresses" table
ALTER TABLE "customer_delivery_addresses" DROP COLUMN "longitude", DROP COLUMN "latitude";
