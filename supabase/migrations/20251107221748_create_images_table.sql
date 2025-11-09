-- Create images table
CREATE TABLE images (
  id BIGSERIAL PRIMARY KEY,
  s3_key TEXT NOT NULL,
  file_url TEXT NOT NULL,
  uploaded_at TIMESTAMP WITH TIME ZONE NOT NULL,
  is_new BOOLEAN NOT NULL DEFAULT true,
  is_displayed BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create unique index on s3_key
CREATE UNIQUE INDEX idx_images_s3_key ON images(s3_key);

-- Create index on is_new for efficient querying of new images
CREATE INDEX idx_images_is_new ON images(is_new);

-- Create index on is_displayed for efficient querying of displayed images
CREATE INDEX idx_images_is_displayed ON images(is_displayed);

-- Add updated_at trigger
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_images_updated_at BEFORE UPDATE ON images
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Enable Row Level Security (RLS)
ALTER TABLE images ENABLE ROW LEVEL SECURITY;

-- Grant full access to service_role
CREATE POLICY "service_role_full_access" ON images
FOR ALL
TO service_role
USING (true)
WITH CHECK (true);
