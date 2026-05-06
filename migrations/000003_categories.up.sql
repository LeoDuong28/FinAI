-- Categories (hierarchical)
CREATE TABLE categories (
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name      VARCHAR(100) NOT NULL,
    slug      VARCHAR(100) UNIQUE NOT NULL,
    icon      VARCHAR(50),
    color     VARCHAR(7),
    parent_id UUID REFERENCES categories(id),
    is_system BOOLEAN DEFAULT TRUE
);
CREATE INDEX idx_categories_parent ON categories(parent_id);
CREATE INDEX idx_categories_slug ON categories(slug);

-- Seed default categories
INSERT INTO categories (name, slug, icon, color, parent_id, is_system) VALUES
-- Parent categories
('Income', 'income', 'banknotes', '#22c55e', NULL, TRUE),
('Housing', 'housing', 'home', '#3b82f6', NULL, TRUE),
('Transportation', 'transportation', 'car', '#8b5cf6', NULL, TRUE),
('Food', 'food', 'utensils', '#f97316', NULL, TRUE),
('Shopping', 'shopping', 'shopping-bag', '#ec4899', NULL, TRUE),
('Entertainment', 'entertainment', 'film', '#a855f7', NULL, TRUE),
('Health', 'health', 'heart-pulse', '#ef4444', NULL, TRUE),
('Financial', 'financial', 'landmark', '#6b7280', NULL, TRUE),
('Education', 'education', 'graduation-cap', '#0ea5e9', NULL, TRUE),
('Travel', 'travel', 'plane', '#14b8a6', NULL, TRUE);

-- Child categories (Income)
INSERT INTO categories (name, slug, icon, color, parent_id, is_system)
SELECT 'Salary', 'salary', 'briefcase', '#22c55e', id, TRUE FROM categories WHERE slug = 'income'
UNION ALL SELECT 'Freelance', 'freelance', 'laptop', '#22c55e', id, TRUE FROM categories WHERE slug = 'income'
UNION ALL SELECT 'Investments', 'investments', 'trending-up', '#22c55e', id, TRUE FROM categories WHERE slug = 'income'
UNION ALL SELECT 'Refunds', 'refunds', 'rotate-ccw', '#22c55e', id, TRUE FROM categories WHERE slug = 'income'
UNION ALL SELECT 'Other Income', 'other-income', 'plus-circle', '#22c55e', id, TRUE FROM categories WHERE slug = 'income';

-- Child categories (Housing)
INSERT INTO categories (name, slug, icon, color, parent_id, is_system)
SELECT 'Rent/Mortgage', 'rent-mortgage', 'building', '#3b82f6', id, TRUE FROM categories WHERE slug = 'housing'
UNION ALL SELECT 'Utilities', 'utilities', 'zap', '#3b82f6', id, TRUE FROM categories WHERE slug = 'housing'
UNION ALL SELECT 'Home Insurance', 'home-insurance', 'shield', '#3b82f6', id, TRUE FROM categories WHERE slug = 'housing'
UNION ALL SELECT 'Maintenance', 'maintenance', 'wrench', '#3b82f6', id, TRUE FROM categories WHERE slug = 'housing';

-- Child categories (Transportation)
INSERT INTO categories (name, slug, icon, color, parent_id, is_system)
SELECT 'Gas', 'gas', 'fuel', '#8b5cf6', id, TRUE FROM categories WHERE slug = 'transportation'
UNION ALL SELECT 'Public Transit', 'public-transit', 'train', '#8b5cf6', id, TRUE FROM categories WHERE slug = 'transportation'
UNION ALL SELECT 'Ride Share', 'ride-share', 'car', '#8b5cf6', id, TRUE FROM categories WHERE slug = 'transportation'
UNION ALL SELECT 'Parking', 'parking', 'square-parking', '#8b5cf6', id, TRUE FROM categories WHERE slug = 'transportation'
UNION ALL SELECT 'Car Payment', 'car-payment', 'car', '#8b5cf6', id, TRUE FROM categories WHERE slug = 'transportation';

-- Child categories (Food)
INSERT INTO categories (name, slug, icon, color, parent_id, is_system)
SELECT 'Groceries', 'groceries', 'shopping-cart', '#f97316', id, TRUE FROM categories WHERE slug = 'food'
UNION ALL SELECT 'Restaurants', 'restaurants', 'utensils', '#f97316', id, TRUE FROM categories WHERE slug = 'food'
UNION ALL SELECT 'Coffee', 'coffee', 'coffee', '#f97316', id, TRUE FROM categories WHERE slug = 'food'
UNION ALL SELECT 'Delivery', 'delivery', 'truck', '#f97316', id, TRUE FROM categories WHERE slug = 'food';

-- Child categories (Shopping)
INSERT INTO categories (name, slug, icon, color, parent_id, is_system)
SELECT 'Clothing', 'clothing', 'shirt', '#ec4899', id, TRUE FROM categories WHERE slug = 'shopping'
UNION ALL SELECT 'Electronics', 'electronics', 'smartphone', '#ec4899', id, TRUE FROM categories WHERE slug = 'shopping'
UNION ALL SELECT 'Home Goods', 'home-goods', 'sofa', '#ec4899', id, TRUE FROM categories WHERE slug = 'shopping'
UNION ALL SELECT 'Personal Care', 'personal-care', 'sparkles', '#ec4899', id, TRUE FROM categories WHERE slug = 'shopping';

-- Child categories (Entertainment)
INSERT INTO categories (name, slug, icon, color, parent_id, is_system)
SELECT 'Streaming', 'streaming', 'tv', '#a855f7', id, TRUE FROM categories WHERE slug = 'entertainment'
UNION ALL SELECT 'Gaming', 'gaming', 'gamepad-2', '#a855f7', id, TRUE FROM categories WHERE slug = 'entertainment'
UNION ALL SELECT 'Events', 'events', 'ticket', '#a855f7', id, TRUE FROM categories WHERE slug = 'entertainment'
UNION ALL SELECT 'Hobbies', 'hobbies', 'palette', '#a855f7', id, TRUE FROM categories WHERE slug = 'entertainment';

-- Child categories (Health)
INSERT INTO categories (name, slug, icon, color, parent_id, is_system)
SELECT 'Medical', 'medical', 'stethoscope', '#ef4444', id, TRUE FROM categories WHERE slug = 'health'
UNION ALL SELECT 'Pharmacy', 'pharmacy', 'pill', '#ef4444', id, TRUE FROM categories WHERE slug = 'health'
UNION ALL SELECT 'Gym', 'gym', 'dumbbell', '#ef4444', id, TRUE FROM categories WHERE slug = 'health'
UNION ALL SELECT 'Health Insurance', 'health-insurance', 'shield-plus', '#ef4444', id, TRUE FROM categories WHERE slug = 'health';

-- Financial subcategories
INSERT INTO categories (name, slug, icon, color, parent_id, is_system)
SELECT 'Bank Fees', 'bank-fees', 'alert-circle', '#6b7280', id, TRUE FROM categories WHERE slug = 'financial'
UNION ALL SELECT 'Interest', 'interest', 'percent', '#6b7280', id, TRUE FROM categories WHERE slug = 'financial'
UNION ALL SELECT 'Transfers', 'transfers', 'arrow-left-right', '#6b7280', id, TRUE FROM categories WHERE slug = 'financial';

-- Education subcategories
INSERT INTO categories (name, slug, icon, color, parent_id, is_system)
SELECT 'Tuition', 'tuition', 'school', '#0ea5e9', id, TRUE FROM categories WHERE slug = 'education'
UNION ALL SELECT 'Books', 'books', 'book-open', '#0ea5e9', id, TRUE FROM categories WHERE slug = 'education'
UNION ALL SELECT 'Courses', 'courses', 'monitor', '#0ea5e9', id, TRUE FROM categories WHERE slug = 'education';

-- Travel subcategories
INSERT INTO categories (name, slug, icon, color, parent_id, is_system)
SELECT 'Flights', 'flights', 'plane', '#14b8a6', id, TRUE FROM categories WHERE slug = 'travel'
UNION ALL SELECT 'Hotels', 'hotels', 'bed', '#14b8a6', id, TRUE FROM categories WHERE slug = 'travel'
UNION ALL SELECT 'Vacation', 'vacation', 'palm-tree', '#14b8a6', id, TRUE FROM categories WHERE slug = 'travel';
