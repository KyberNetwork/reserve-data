ALTER TABLE "assets" 
ADD COLUMN normal_update_per_period FLOAT DEFAULT 1 
CHECK(normal_update_per_period > 0),
ADD COLUMN max_imbalance_ratio FLOAT DEFAULT 2
CHECK(max_imbalance_ratio > 0);
