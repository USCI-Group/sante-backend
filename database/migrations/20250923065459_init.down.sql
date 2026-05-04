-- reverse: create "feedbacks" table
DROP TABLE "feedbacks";
-- reverse: modify "orders" table
ALTER TABLE "orders" DROP COLUMN "payment_channel";
-- reverse: modify "customers" table
ALTER TABLE "customers" DROP COLUMN "profile_picture";
