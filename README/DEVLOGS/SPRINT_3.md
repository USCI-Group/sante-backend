
## SPRINT 3 DEVLOG (6/10/2025 - 16/10/2025)

### Order Processing (Backend/Frontend)
- Implemented API endpoints for order creation, validation, payment initiation, and order status updates.
- Developed frontend flow to guide the user through cart, address selection, payment, and confirmation screens.
- Added validation for stock availability, product status, and business hour checks before allowing order placement.

### Payment Processing (Backend/Frontend)
- Integrated third-party payment gateway APIs for handling payments.
- Developed backend logic for payment initiation, callback webhook handling, and secure payment confirmation.
- On frontend, built user flows for payment method selection, redirection to payment provider, and error handling for failed transactions.

### Reorder Process (Backend/Frontend)
- Backend: Added API to fetch last successful order for quick reordering.
- Enabled duplication of previous order details including products, modifiers, and delivery preferences.
- Frontend: Created "Reorder" button in order history view for fast checkout.

### Delivery Address Process (Backend/Frontend)
- Backend: Designed endpoints for address CRUD operations tied to a user profile.
- Added address validation (postcode, state, etc.) with automatic suggestions.
- Frontend: Implemented address management UI. Integrated autocomplete for locality, state, and postal code.

### Order History (Backend/Frontend)
- Backend APIs for fetching order history with filter and pagination support.
- Frontend UI to display order summary, details, and reordering option.

### Dynamic Payment Configuration (Backend/Frontend)
- Backend: Added configuration tables/models for dynamic payment methods by business/operator.
- Enabled toggling visibility of payment methods based on business settings.
- Frontend: Dynamically renders available payment methods for user per business during checkout.

---

### Membership Loyalty Management Features (Web Admin, Loyalty App)

#### Manage Rank Tiers (Web Admin)
- Developed UI to add, update, and order customer rank tiers (e.g. Silver, Gold, Platinum).
- Backend updates to support querying and modifying tier hierarchy.
  
#### Rank Tiers: Manage Tier Rules (Web Admin)
- Created interfaces for defining rules for tier attainment and retention (e.g. spend, orders count).
- Backend validation for rule logic integrity and prevention of overlapping rules.

#### Rank Tiers: Manage Benefits (Web Admin)
- UI for associating benefits (e.g. discounts, exclusive access) to tiers.
- Backend relationships ensured tier-benefit mappings are unique and effective-dated.

#### Manage Vouchers (Web Admin)
- Implemented comprehensive voucher CRUD (create, update, delete) screens.
- Backend improvements for voucher distribution validity, application rules, and redemption tracking.

#### Manage Missions (Web Admin)
- UI for creating and managing missions including conditions and rewards.
- Backend: API endpoints for mission CRUD operations, with validation for overlapping and duplicate missions.

#### Missions: Add Completion Criteria (Web Admin)
- Added UI/logic for mission criteria such as product purchase, order value, outlet visit, and membership-based criteria.
- Ensured backend supports flexible definition of criteria with robust validation.

#### Missions: Add Completion Rewards (Web Admin)
- Extended mission creation with reward types (points, vouchers, etc.).
- UI and backend supports structured reward assignment and editing.

---

### Loyalty App (Pretzely) API/UI Integrations

#### Voucher Membership UI and Functionality API
- Consumed voucher-related endpoints: display available/activated vouchers, redemption status.
- UI improvements for ease of voucher activation/redemption and sorting/filtering.

#### Mission Membership UI and Functionality API
- Integrated with mission APIs to display ongoing/completed missions, rules, rewards.
- Added progress tracking visuals and claim completion flows.

---

### Crash Analytics & Distribution

#### Firebase CrashAnalytics and App Distribution (Loyalty App)
- Configured Firebase Crashlytics for real-time error tracking.
- Automated build upload to Firebase App Distribution for testing and UAT releases.

---


