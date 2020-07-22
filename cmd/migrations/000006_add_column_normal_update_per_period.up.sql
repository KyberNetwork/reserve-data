ALTER TABLE "assets" 
ADD COLUMN IF NOT EXISTS normal_update_per_period FLOAT DEFAULT 1 
CHECK(normal_update_per_period > 0);
