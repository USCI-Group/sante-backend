-- modify "business_configurations" table
ALTER TABLE "business_configurations" ADD COLUMN "privacy_policy" text NULL;
-- modify "customer_membership_stats" table
ALTER TABLE "customer_membership_stats" ADD COLUMN "total_experience_points" bigint NULL DEFAULT 0;
-- modify "products" table
ALTER TABLE "products" ADD COLUMN "experience_points" integer NULL DEFAULT 0;
