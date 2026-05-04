-- reverse: modify "products" table
ALTER TABLE "products" DROP COLUMN "experience_points";
-- reverse: modify "customer_membership_stats" table
ALTER TABLE "customer_membership_stats" DROP COLUMN "total_experience_points";
-- reverse: modify "business_configurations" table
ALTER TABLE "business_configurations" DROP COLUMN "privacy_policy";
