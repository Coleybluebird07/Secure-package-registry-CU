CREATE TYPE REVIEW_STATUS AS ENUM ('pending', 'approved', 'rejected');

CREATE TABLE package_reviews (
                                 id SERIAL PRIMARY KEY,
                                 package_version_id INTEGER NOT NULL REFERENCES package_versions(id) ON DELETE CASCADE,
                                 status REVIEW_STATUS NOT NULL DEFAULT 'pending',
                                 notes TEXT,
                                 reviewed_by text REFERENCES "user"(id),
                                 reviewed_at TIMESTAMP WITH TIME ZONE,
                                 created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                 updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                 UNIQUE(package_version_id)
);