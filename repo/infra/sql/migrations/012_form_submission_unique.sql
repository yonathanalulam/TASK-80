
CREATE UNIQUE INDEX IF NOT EXISTS uq_form_submissions_itinerary_member
    ON itinerary_member_form_submissions (itinerary_id, member_user_id);
