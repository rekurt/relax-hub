INSERT INTO platform_settings (key, value, description, type) VALUES
    ('listing_wizard_video_url', '', 'URL for the listing creation wizard intro video (YouTube/Vimeo embed or direct link)', 'string')
ON CONFLICT (key) DO NOTHING;
