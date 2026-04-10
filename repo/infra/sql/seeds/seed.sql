
INSERT INTO permissions (id, resource, action, description) VALUES
  (gen_random_uuid(), 'itineraries', 'create', 'Create itineraries'),
  (gen_random_uuid(), 'itineraries', 'read', 'View itineraries'),
  (gen_random_uuid(), 'itineraries', 'update', 'Edit itineraries'),
  (gen_random_uuid(), 'itineraries', 'delete', 'Delete itineraries'),
  (gen_random_uuid(), 'itineraries', 'publish', 'Publish itineraries'),
  (gen_random_uuid(), 'bookings', 'create', 'Create bookings'),
  (gen_random_uuid(), 'bookings', 'read', 'View bookings'),
  (gen_random_uuid(), 'bookings', 'checkout', 'Checkout bookings'),
  (gen_random_uuid(), 'procurement', 'create_rfq', 'Create RFQs'),
  (gen_random_uuid(), 'procurement', 'quote', 'Submit quotes'),
  (gen_random_uuid(), 'procurement', 'create_po', 'Create purchase orders'),
  (gen_random_uuid(), 'procurement', 'receive', 'Receive deliveries'),
  (gen_random_uuid(), 'procurement', 'inspect', 'Perform inspections'),
  (gen_random_uuid(), 'finance', 'record_tender', 'Record tenders'),
  (gen_random_uuid(), 'finance', 'approve_refund', 'Approve refunds'),
  (gen_random_uuid(), 'finance', 'process_withdrawal', 'Process withdrawals'),
  (gen_random_uuid(), 'finance', 'view_ledger', 'View ledger'),
  (gen_random_uuid(), 'finance', 'reconcile', 'Run reconciliation'),
  (gen_random_uuid(), 'files', 'upload', 'Upload files'),
  (gen_random_uuid(), 'files', 'download', 'Download files'),
  (gen_random_uuid(), 'contracts', 'generate', 'Generate contracts'),
  (gen_random_uuid(), 'invoices', 'request', 'Request invoices'),
  (gen_random_uuid(), 'invoices', 'generate', 'Generate invoices'),
  (gen_random_uuid(), 'reviews', 'create', 'Submit reviews'),
  (gen_random_uuid(), 'reviews', 'read', 'View reviews'),
  (gen_random_uuid(), 'admin', 'manage_users', 'Manage users'),
  (gen_random_uuid(), 'admin', 'manage_coupons', 'Manage coupons'),
  (gen_random_uuid(), 'admin', 'view_audit', 'View audit logs'),
  (gen_random_uuid(), 'admin', 'manage_blacklist', 'Manage blacklists'),
  (gen_random_uuid(), 'admin', 'approve_actions', 'Approve risk-throttled actions'),
  (gen_random_uuid(), 'notifications', 'read', 'View notifications'),
  (gen_random_uuid(), 'notifications', 'manage', 'Manage notifications'),
  (gen_random_uuid(), 'wallets', 'view_own', 'View own wallet'),
  (gen_random_uuid(), 'wallets', 'request_withdrawal', 'Request withdrawal'),
  (gen_random_uuid(), 'members', 'submit_form', 'Submit member forms')
;


INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'traveler' AND (p.resource, p.action) IN (
  ('itineraries','read'), ('bookings','read'), ('files','download'), ('invoices','request'),
  ('reviews','create'), ('reviews','read'), ('notifications','read'), ('wallets','view_own'),
  ('members','submit_form'), ('files','upload')
) ;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'group_organizer' AND (p.resource, p.action) IN (
  ('itineraries','create'), ('itineraries','read'), ('itineraries','update'), ('itineraries','publish'),
  ('bookings','create'), ('bookings','read'), ('bookings','checkout'),
  ('procurement','create_rfq'), ('procurement','create_po'),
  ('files','upload'), ('files','download'), ('invoices','request'),
  ('reviews','create'), ('reviews','read'),
  ('notifications','read'), ('notifications','manage'), ('wallets','view_own')
) ;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'supplier' AND (p.resource, p.action) IN (
  ('procurement','quote'), ('procurement','receive'),
  ('files','upload'), ('files','download'),
  ('reviews','create'), ('reviews','read'),
  ('notifications','read'), ('wallets','view_own')
) ;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'courier_runner' AND (p.resource, p.action) IN (
  ('procurement','receive'), ('files','upload'), ('files','download'),
  ('reviews','create'), ('reviews','read'),
  ('notifications','read'), ('wallets','view_own'), ('wallets','request_withdrawal')
) ;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'accountant' AND (p.resource, p.action) IN (
  ('bookings','read'), ('finance','record_tender'), ('finance','approve_refund'),
  ('finance','process_withdrawal'), ('finance','view_ledger'), ('finance','reconcile'),
  ('files','download'), ('invoices','generate'), ('notifications','read')
) ;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'administrator'
;

INSERT INTO users (id, email, password_hash, status) VALUES
  ('c0000000-0000-0000-0000-000000000001', 'admin@travel.local',      '$2a$10$tjWtCQ.BCUMjSN9/pj5xbucuNoeTfetO9Bc94IYiSJBDanpovkwVO', 'active'),
  ('c0000000-0000-0000-0000-000000000002', 'organizer1@travel.local',  '$2a$10$tjWtCQ.BCUMjSN9/pj5xbucuNoeTfetO9Bc94IYiSJBDanpovkwVO', 'active'),
  ('c0000000-0000-0000-0000-000000000003', 'organizer2@travel.local',  '$2a$10$tjWtCQ.BCUMjSN9/pj5xbucuNoeTfetO9Bc94IYiSJBDanpovkwVO', 'active'),
  ('c0000000-0000-0000-0000-000000000004', 'traveler1@travel.local',   '$2a$10$tjWtCQ.BCUMjSN9/pj5xbucuNoeTfetO9Bc94IYiSJBDanpovkwVO', 'active'),
  ('c0000000-0000-0000-0000-000000000005', 'traveler2@travel.local',   '$2a$10$tjWtCQ.BCUMjSN9/pj5xbucuNoeTfetO9Bc94IYiSJBDanpovkwVO', 'active'),
  ('c0000000-0000-0000-0000-000000000006', 'traveler3@travel.local',   '$2a$10$tjWtCQ.BCUMjSN9/pj5xbucuNoeTfetO9Bc94IYiSJBDanpovkwVO', 'active'),
  ('c0000000-0000-0000-0000-000000000007', 'traveler4@travel.local',   '$2a$10$tjWtCQ.BCUMjSN9/pj5xbucuNoeTfetO9Bc94IYiSJBDanpovkwVO', 'active'),
  ('c0000000-0000-0000-0000-000000000008', 'supplier1@travel.local',   '$2a$10$tjWtCQ.BCUMjSN9/pj5xbucuNoeTfetO9Bc94IYiSJBDanpovkwVO', 'active'),
  ('c0000000-0000-0000-0000-000000000009', 'supplier2@travel.local',   '$2a$10$tjWtCQ.BCUMjSN9/pj5xbucuNoeTfetO9Bc94IYiSJBDanpovkwVO', 'active'),
  ('c0000000-0000-0000-0000-000000000010', 'supplier3@travel.local',   '$2a$10$tjWtCQ.BCUMjSN9/pj5xbucuNoeTfetO9Bc94IYiSJBDanpovkwVO', 'active'),
  ('c0000000-0000-0000-0000-000000000011', 'courier1@travel.local',    '$2a$10$tjWtCQ.BCUMjSN9/pj5xbucuNoeTfetO9Bc94IYiSJBDanpovkwVO', 'active'),
  ('c0000000-0000-0000-0000-000000000012', 'courier2@travel.local',    '$2a$10$tjWtCQ.BCUMjSN9/pj5xbucuNoeTfetO9Bc94IYiSJBDanpovkwVO', 'active'),
  ('c0000000-0000-0000-0000-000000000013', 'accountant@travel.local',  '$2a$10$tjWtCQ.BCUMjSN9/pj5xbucuNoeTfetO9Bc94IYiSJBDanpovkwVO', 'active')
;

INSERT INTO user_roles (user_id, role_id) SELECT 'c0000000-0000-0000-0000-000000000001', id FROM roles WHERE name = 'administrator' ;
INSERT INTO user_roles (user_id, role_id) SELECT 'c0000000-0000-0000-0000-000000000002', id FROM roles WHERE name = 'group_organizer' ;
INSERT INTO user_roles (user_id, role_id) SELECT 'c0000000-0000-0000-0000-000000000003', id FROM roles WHERE name = 'group_organizer' ;
INSERT INTO user_roles (user_id, role_id) SELECT 'c0000000-0000-0000-0000-000000000004', id FROM roles WHERE name = 'traveler' ;
INSERT INTO user_roles (user_id, role_id) SELECT 'c0000000-0000-0000-0000-000000000005', id FROM roles WHERE name = 'traveler' ;
INSERT INTO user_roles (user_id, role_id) SELECT 'c0000000-0000-0000-0000-000000000006', id FROM roles WHERE name = 'traveler' ;
INSERT INTO user_roles (user_id, role_id) SELECT 'c0000000-0000-0000-0000-000000000007', id FROM roles WHERE name = 'traveler' ;
INSERT INTO user_roles (user_id, role_id) SELECT 'c0000000-0000-0000-0000-000000000008', id FROM roles WHERE name = 'supplier' ;
INSERT INTO user_roles (user_id, role_id) SELECT 'c0000000-0000-0000-0000-000000000009', id FROM roles WHERE name = 'supplier' ;
INSERT INTO user_roles (user_id, role_id) SELECT 'c0000000-0000-0000-0000-000000000010', id FROM roles WHERE name = 'supplier' ;
INSERT INTO user_roles (user_id, role_id) SELECT 'c0000000-0000-0000-0000-000000000011', id FROM roles WHERE name = 'courier_runner' ;
INSERT INTO user_roles (user_id, role_id) SELECT 'c0000000-0000-0000-0000-000000000012', id FROM roles WHERE name = 'courier_runner' ;
INSERT INTO user_roles (user_id, role_id) SELECT 'c0000000-0000-0000-0000-000000000013', id FROM roles WHERE name = 'accountant' ;

INSERT INTO user_profiles (user_id, display_name) VALUES
  ('c0000000-0000-0000-0000-000000000001', 'Admin User'),
  ('c0000000-0000-0000-0000-000000000002', 'Sarah Organizer'),
  ('c0000000-0000-0000-0000-000000000003', 'Mike Organizer'),
  ('c0000000-0000-0000-0000-000000000004', 'Alice Traveler'),
  ('c0000000-0000-0000-0000-000000000005', 'Bob Traveler'),
  ('c0000000-0000-0000-0000-000000000006', 'Carol Traveler'),
  ('c0000000-0000-0000-0000-000000000007', 'Dave Traveler'),
  ('c0000000-0000-0000-0000-000000000008', 'Summit Hotels'),
  ('c0000000-0000-0000-0000-000000000009', 'Pacific Transport Co'),
  ('c0000000-0000-0000-0000-000000000010', 'Mountain Gear Supply'),
  ('c0000000-0000-0000-0000-000000000011', 'Express Runner Jake'),
  ('c0000000-0000-0000-0000-000000000012', 'Swift Delivery Lisa'),
  ('c0000000-0000-0000-0000-000000000013', 'Finance Team')
;

INSERT INTO do_not_disturb_settings (user_id, dnd_start, dnd_end, enabled) VALUES
  ('c0000000-0000-0000-0000-000000000004', '21:00', '08:00', true),
  ('c0000000-0000-0000-0000-000000000005', '21:00', '08:00', true),
  ('c0000000-0000-0000-0000-000000000006', '22:00', '07:00', true)
;

INSERT INTO itineraries (id, organizer_id, title, meetup_at, meetup_location_text, notes, status, published_at) VALUES
  ('d0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000002',
   'Mountain Trail Adventure 2026',
   '2026-07-14 18:30:00+00', 'Central Station Main Hall, Platform 3',
   'Pack light, bring rain gear. Altitude may reach 3000m. Group dinner on Day 1.',
   'published', '2026-04-01 10:00:00+00')
;

INSERT INTO itinerary_checkpoints (id, itinerary_id, sort_order, checkpoint_text, eta) VALUES
  ('e0000000-0000-0000-0000-000000000001', 'd0000000-0000-0000-0000-000000000001', 1, 'Depart Central Station', '2026-07-14 18:30:00+00'),
  ('e0000000-0000-0000-0000-000000000002', 'd0000000-0000-0000-0000-000000000001', 2, 'Arrive Base Camp Lodge', '2026-07-14 22:00:00+00'),
  ('e0000000-0000-0000-0000-000000000003', 'd0000000-0000-0000-0000-000000000001', 3, 'Summit Trail Start', '2026-07-15 06:00:00+00'),
  ('e0000000-0000-0000-0000-000000000004', 'd0000000-0000-0000-0000-000000000001', 4, 'Ridge Viewpoint Rest', '2026-07-15 12:00:00+00'),
  ('e0000000-0000-0000-0000-000000000005', 'd0000000-0000-0000-0000-000000000001', 5, 'Return to Lodge', '2026-07-15 17:00:00+00')
;

INSERT INTO itinerary_members (id, itinerary_id, user_id, role) VALUES
  ('f0000000-0000-0000-0000-000000000001', 'd0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000004', 'participant'),
  ('f0000000-0000-0000-0000-000000000002', 'd0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000005', 'participant'),
  ('f0000000-0000-0000-0000-000000000003', 'd0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000006', 'participant')
;

INSERT INTO itinerary_member_form_definitions (id, itinerary_id, field_key, field_label, field_type, required, options_json, sort_order, active) VALUES
  ('f1000000-0000-0000-0000-000000000001', 'd0000000-0000-0000-0000-000000000001', 'vehicle_plate', 'Vehicle Plate Number', 'text', false, null, 1, true),
  ('f1000000-0000-0000-0000-000000000002', 'd0000000-0000-0000-0000-000000000001', 'emergency_contact', 'Emergency Contact Name', 'text', true, null, 2, true),
  ('f1000000-0000-0000-0000-000000000003', 'd0000000-0000-0000-0000-000000000001', 'emergency_phone', 'Emergency Contact Phone', 'text', true, null, 3, true),
  ('f1000000-0000-0000-0000-000000000004', 'd0000000-0000-0000-0000-000000000001', 'dietary', 'Dietary Requirements', 'select', false, '["None","Vegetarian","Vegan","Halal","Kosher","Gluten-Free","Other"]', 4, true),
  ('f1000000-0000-0000-0000-000000000005', 'd0000000-0000-0000-0000-000000000001', 'medical_notes', 'Medical Notes (Optional)', 'textarea', false, null, 5, true)
;

INSERT INTO coupons (id, code, name, discount_type, amount, min_spend, percent_off, valid_from, valid_to, eligibility_json, stack_group, exclusive, usage_limit_total, usage_limit_per_user, active) VALUES
  ('d1000000-0000-0000-0000-000000000001', 'SAVE25', '$25 Off Orders Over $200', 'threshold_fixed', 25.00, 200.00, NULL, '2026-01-01 00:00:00+00', '2026-12-31 23:59:59+00', '{"categories":["lodging","transport","activity"]}', 'threshold', false, 1000, 3, true),
  ('d1000000-0000-0000-0000-000000000002', 'LODGE10', '10% Off Lodging', 'percentage', NULL, NULL, 10.00, '2026-01-01 00:00:00+00', '2026-12-31 23:59:59+00', '{"categories":["lodging"]}', 'percentage', false, 500, 2, true),
  ('d1000000-0000-0000-0000-000000000003', 'WELCOME15', 'Welcome Gift - $15 Off', 'new_user_gift', 15.00, 50.00, NULL, '2026-01-01 00:00:00+00', '2026-12-31 23:59:59+00', '{"new_user_only":true,"validity_days_from_signup":14}', 'new_user', true, NULL, 1, true)
;

INSERT INTO bookings (id, organizer_id, itinerary_id, title, description, status, total_amount, discount_amount, escrow_amount) VALUES
  ('d3000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000002', 'd0000000-0000-0000-0000-000000000001',
   'Mountain Adventure Lodging & Transport', 'Base Camp Lodge 2-night stay + group transport',
   'draft', 850.00, 0.00, 0.00)
;

INSERT INTO booking_items (id, booking_id, item_type, item_name, description, unit_price, quantity, subtotal, category) VALUES
  ('d4000000-0000-0000-0000-000000000001', 'd3000000-0000-0000-0000-000000000001', 'lodging', 'Base Camp Lodge - Double Room', '2 nights, breakfast included', 150.00, 3, 450.00, 'lodging'),
  ('d4000000-0000-0000-0000-000000000002', 'd3000000-0000-0000-0000-000000000001', 'transport', 'Group Shuttle - Central to Base Camp', 'Round trip, 15-seater', 200.00, 1, 200.00, 'transport'),
  ('d4000000-0000-0000-0000-000000000003', 'd3000000-0000-0000-0000-000000000001', 'activity', 'Guided Summit Trail Tour', 'Professional guide, safety equipment', 50.00, 4, 200.00, 'activity')
;

INSERT INTO wallets (id, owner_id, wallet_type, balance, currency) VALUES
  ('d5000000-0000-0000-0000-000000000001', NULL, 'platform_clearing', 0.00, 'USD'),
  ('d5000000-0000-0000-0000-000000000002', NULL, 'escrow_control', 0.00, 'USD'),
  ('d5000000-0000-0000-0000-000000000003', NULL, 'refund_clearing', 0.00, 'USD'),
  ('d5000000-0000-0000-0000-000000000004', NULL, 'fee_revenue', 0.00, 'USD'),
  ('d5000000-0000-0000-0000-000000000005', 'c0000000-0000-0000-0000-000000000008', 'supplier', 0.00, 'USD'),
  ('d5000000-0000-0000-0000-000000000006', 'c0000000-0000-0000-0000-000000000009', 'supplier', 0.00, 'USD'),
  ('d5000000-0000-0000-0000-000000000007', 'c0000000-0000-0000-0000-000000000011', 'courier', 0.00, 'USD'),
  ('d5000000-0000-0000-0000-000000000008', 'c0000000-0000-0000-0000-000000000012', 'courier', 0.00, 'USD')
;

INSERT INTO review_dimensions (id, name, label, active) VALUES
  (gen_random_uuid(), 'punctuality', 'Punctuality', true),
  (gen_random_uuid(), 'communication', 'Communication', true),
  (gen_random_uuid(), 'quality', 'Quality', true),
  (gen_random_uuid(), 'compliance', 'Compliance', true),
  (gen_random_uuid(), 'professionalism', 'Professionalism', true),
  (gen_random_uuid(), 'accuracy', 'Accuracy', true),
  (gen_random_uuid(), 'cleanliness', 'Cleanliness', true),
  (gen_random_uuid(), 'route_adherence', 'Route/Service Adherence', true),
  (gen_random_uuid(), 'delivery_integrity', 'Delivery Integrity', true)
;

INSERT INTO credit_tiers (id, tier_name, min_transactions, min_avg_rating, max_violations, description) VALUES
  (gen_random_uuid(), 'bronze', 0, 0.0, 999, 'Default starting tier'),
  (gen_random_uuid(), 'silver', 5, 3.5, 5, 'Established participant'),
  (gen_random_uuid(), 'gold', 20, 4.0, 2, 'Trusted participant'),
  (gen_random_uuid(), 'platinum', 50, 4.5, 0, 'Elite participant'),
  (gen_random_uuid(), 'restricted', 0, 0.0, 0, 'Account under restriction')
;

INSERT INTO message_templates (id, template_key, subject_template, body_template, channel_type, active) VALUES
  (gen_random_uuid(), 'itinerary_published', 'Itinerary Published: {{title}}', 'The itinerary "{{title}}" has been published by {{organizer}}. Meetup: {{meetup_at}} at {{location}}.', 'in_app', true),
  (gen_random_uuid(), 'itinerary_changed', 'Itinerary Updated: {{title}}', '{{organizer}} made changes to "{{title}}": {{change_summary}}', 'in_app', true),
  (gen_random_uuid(), 'booking_confirmed', 'Booking Confirmed: {{title}}', 'Your booking "{{title}}" has been confirmed. Total: ${{total}}. Escrow held.', 'in_app', true),
  (gen_random_uuid(), 'rfq_issued', 'New RFQ: {{title}}', 'You have been invited to quote on "{{title}}". Deadline: {{deadline}}.', 'in_app', true),
  (gen_random_uuid(), 'withdrawal_approved', 'Withdrawal Approved', 'Your withdrawal request of ${{amount}} has been approved and will be settled.', 'in_app', true),
  (gen_random_uuid(), 'inspection_failed', 'Inspection Failed: PO {{po_number}}', 'Inspection for PO {{po_number}} has failed. Notes: {{notes}}. A discrepancy ticket has been created.', 'in_app', true)
;

INSERT INTO contract_templates (id, name, body_template, variable_schema_json, active, version) VALUES
  (gen_random_uuid(), 'Supplier Service Agreement',
   'SERVICE AGREEMENT

This agreement is entered between {{platform_name}} and {{supplier_name}} on {{effective_date}}.

SCOPE OF SERVICES:
{{service_description}}

DATES: {{service_start}} to {{service_end}}
TOTAL VALUE: ${{total_amount}}
PAYMENT TERMS: {{payment_terms}}

OBLIGATIONS:
1. Supplier shall deliver services as specified.
2. Quality standards as per procurement specifications.
3. Compliance with all applicable local regulations.

SIGNED:
Platform Representative: _______________
Supplier Representative: _______________
Date: {{signature_date}}',
   '{"properties": {"platform_name": {"type": "string"}, "supplier_name": {"type": "string"}, "effective_date": {"type": "string"}, "service_description": {"type": "string"}, "service_start": {"type": "string"}, "service_end": {"type": "string"}, "total_amount": {"type": "string"}, "payment_terms": {"type": "string"}, "signature_date": {"type": "string"}}}',
   true, 1)
;
