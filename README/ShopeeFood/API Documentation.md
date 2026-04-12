# ShopeeFood API endpoints documentation

This document outlines the detailed step-by-step flow for all ShopeeFood API endpoints in our system, covering authentication, menu operations, and order management.

---

## Table of Contents
1. [OAuth Token Endpoint](#1-oauth-token-endpoint)
2. [Get Food Menu Endpoint](#2-get-food-menu-endpoint)
3. [Menu Sync Notification Result Endpoint](#3-menu-sync-notification-result-endpoint)
4. [Trigger Menu Update Endpoint](#4-trigger-menu-update-endpoint)
5. [Order Submission Endpoint](#5-order-submission-endpoint)
6. [Order Status Update Endpoint](#6-order-status-update-endpoint)
7. [Get Order from Shopee Endpoint](#7-get-order-from-shopee-endpoint)

---

## 1. OAuth Token Endpoint

**Path:** `/api/shopeefood/gettoken`  
**Method:** `POST`  
**Authentication:** `public` (called by Shopee)  
**Handler:** `ShopeeGetPartnerOauthToken`

### Step-by-Step Flow:

1. **Request Reception**
   - Shopee sends POST request with `client_id` and `client_secret` in request body
   - Request payload: `OAuthRequest{ClientID, ClientSecret}`

2. **Request Caching**
   - Create a new `Cache` record in database
   - Cache key: `"shopeefood_get_partner_oauth_token"`
   - Cache value: String representation of the entire request payload
   - Log any errors during caching (non-blocking)

3. **Request Validation**
   - Validate request struct using `govalidator.ValidateStruct()`
   - Ensure `ClientID` and `ClientSecret` are present and not empty
   - Return error if validation fails

4. **Credential Verification**
   - Retrieve system configuration values:
     - `SystemDataInfoTypeGrabFoodPartnerClientID` (note: uses GrabFood constant name)
     - `SystemDataInfoTypeGrabFoodPartnerClientSecret`
   - Compare provided `ClientID` and `ClientSecret` with system values
   - If mismatch, return `Unauthenticated` error with message "Invalid client credentials"

5. **JWT Token Generation**
   - Create JWT token using HS256 signing method
   - Token claims include:
     - `sub`: Partner client ID
     - `exp`: Current time + 24 hours
     - `login_type`: "shopeefood"
   - Sign token using `JWT_SECRET_KEY` from environment

6. **Response**
   - Return `ShopeeOauthResponse`:
     - `access_token`: Generated JWT token string
     - `expires_in`: 86400 seconds (24 hours)

---

## 2. Get Food Menu Endpoint

**Path:** `/api/shopeefood/menu`  
**Method:** `GET`  
**Authentication:** `auth` (called by Shopee)  
**Handler:** `ShopeeGetFoodMenu`

### Step-by-Step Flow:

1. **Request Reception**
   - Shopee sends GET request with `store_id` as query parameter
   - Extract `store_id` from `req.URL.Query().Get("store_id")`

2. **Request Caching**
   - Create a new `Cache` record in database
   - Cache key: `"shopeefood_get_menu"`
   - Cache value: `"store_id: {store_id}"`
   - Log any errors during caching (non-blocking)

3. **Store ID Validation**
   - Check if `store_id` is empty
   - If empty, return HTTP 400 with error message: `{"error": "store_id is required"}`

4. **Merchant Secret Lookup**
   - Query `MerchantSecret` table where `shopee_store_id = store_id`
   - If not found:
     - Log error
     - Return HTTP 404 with `CatalogResponse`:
       - `code`: 500
       - `message`: "Internal server error"
       - `data`: Empty `CatalogsWrapper`

5. **Catalog Generation**
   - Call `getCatalogs(merchantSecret.OutletID)` to build menu structure:
     - **Step 5.1:** Fetch outlet by `outletID`
     - **Step 5.2:** Query products where:
       - `business_id` matches outlet's business
       - `is_shopee_food = true`
       - `is_active = true`
       - Ordered by `sort_order ASC`
     - **Step 5.3:** Fetch `ProductCategoryMapping` records for these products
       - Preload `Product` and `ProductCategory` relationships
     - **Step 5.4:** Sort mappings by product `sort_order`
     - **Step 5.5:** Group mappings by `ProductCategoryID`
     - **Step 5.6:** For each category:
       - Build `Dish` array from products in category
       - Each dish includes:
         - `id`: Product UUID string
         - `name`: Product name
         - `price`: Product price (in cents)
         - `picture`: Product image URL
         - `description`: Product description
         - `available`: Always `true` (SHOPEEFOOD_AVAILABLE constant)
         - `sales_time`: Generated sales time (08:00-18:00, all days)(Hardcoded, probably need update)
         - `option_groups`: Generated from product modifier groups
     - **Step 5.7:** For each product, generate option groups:
       - Query `ProductModifierMapping` for the product
       - Preload `ModifierGroup`
       - For each modifier group:
         - Query active `ModifierOptions` for the group
         - Build `Option` array with:
           - `id`: Option UUID string
           - `name`: Option name
           - `price`: Price adjustment (in cents)
           - `available`: Always `true`
         - Build `OptionGroup` with:
           - `id`: Modifier group UUID string
           - `name`: Modifier group name
           - `select_min`: 1
           - `select_max`: Max selection from mapping
           - `options`: Generated options array
     - **Step 5.8:** Build `Catalog` array with:
       - `id`: Category UUID string
       - `name`: Category name
       - `dishes`: Generated dishes array

6. **Response**
   - Return HTTP 200 with `CatalogResponse`:
     - `code`: 0
     - `message`: "OK"
     - `data`: `CatalogsWrapper` containing catalogs array

---

## 3. Menu Sync Notification Result Endpoint

**Path:** `/api/shopeefood/menu/notification/result`  
**Method:** `POST`  
**Authentication:** `auth` (called by Shopee)  
**Handler:** `ShopeeMenuSync`

### Step-by-Step Flow:

1. **Request Reception**
   - Shopee sends POST request with menu sync result
   - Request payload: `MenuSyncRequest{StoreID, Result, Message, NotificationID, NotificationTime}`

2. **Request Caching**
   - Create a new `Cache` record in database
   - Cache key: `"shopeefood_menu_sync"`
   - Cache value: String representation of entire request
   - Log any errors during caching (non-blocking)

3. **Request Validation**
   - Validate request struct using `govalidator.ValidateStruct()`
   - Ensure `StoreID` is present
   - If validation fails, return error response with code 500

4. **Log Notification Details**
   - Log `NotificationID` (X-SF-RequestID)
   - Log `NotificationTime` (X-SF-Timestamp)

5. **Merchant Secret Lookup**
   - Query `MerchantSecret` table where `shopee_store_id = req.StoreID`
   - If not found:
     - Log error
     - Return error response with code 500 and message "Internal server error"

6. **Result Processing**
   - Check if `req.Result` is nil or not equal to 0
   - **If sync failed (Result != 0 or nil):**
     - Update `merchantSecret.ShopeeMenuSyncState = "FAILED"`
     - Save to database
     - Return response:
       - `code`: `req.Result` value
       - `message`: `"Menu not sync: {req.Message}"`
   - **If sync succeeded (Result == 0):**
     - Update `merchantSecret.ShopeeMenuSyncState = "SUCCESS"`
     - Save to database
     - Return response:
       - `code`: 0
       - `message`: "Menu sync"

7. **Error Handling**
   - If database save fails, log error and return error response with code 500

---

## 4. Trigger Menu Update Endpoint

**Path:** `/api/shopeefood/menu/trigger-update/:outletID`  
**Method:** `GET`  
**Authentication:** `auth` (internal API)  
**Handler:** `TriggerMenuUpdate`

### Step-by-Step Flow:

1. **Request Reception**
   - Internal system sends GET request with `outletID` as path parameter
   - Extract `outletID` from URL path

2. **Outlet Lookup**
   - Query `Outlet` table where `id = outletID`
   - If not found:
     - Log error
     - Return error

3. **Business Configuration Lookup**
   - Query `BusinessConfiguration` table where `business_id = outlet.BusinessID`
   - If not found:
     - Log error
     - Return error

4. **Merchant Secret Lookup**
   - Query `MerchantSecret` table where `outlet_id = outletID`
   - If not found:
     - Log error
     - Return error

5. **Request Body Preparation**
   - Create request body map:
     - `store_id`: `merchantSecret.ShopeeStoreID`

6. **Shopee API URL Construction**
   - Get base URL from `getShopeeBaseURL()` (UAT or Production based on ENV)
   - Get `clientID` from `businessConfiguration.ShopeeClientID`
   - Construct full URL: `{baseURL}/api/adaptor/{clientID}/v1/store/menu/notification`

7. **HTTP Request Creation**
   - Create PUT request to Shopee API
   - Marshal request body to JSON
   - Create `http.Request` with PUT method, URL, and JSON body

8. **Client Secret Decryption**
   - Decrypt `businessConfiguration.ShopeeClientSecret` using `common.DecryptText()`
   - If decryption fails:
     - Log error
     - Return error

9. **Request Header Preparation**
   - Call `headerPrepWithBody()` to prepare Shopee authentication headers:
     - **Step 9.1:** Get Shopee access token by calling `getShopeeToken(businessID)`:
       - Query `BusinessConfiguration` for business
       - Decrypt client secret
       - Create POST request to Shopee token endpoint: `/api/adaptor/{clientID}/v1/auth/token`
       - Set Basic Auth header with base64-encoded `clientID:clientSecret`
       - Set Content-Type: `application/x-www-form-urlencoded`
       - Request body: `{grant_type: "client_credentials", scope: "all"}`
       - Parse response to get `access_token`
     - **Step 9.2:** Generate HMAC-SHA256 signature:
       - Create signature string: `access_token={token}&app_id={clientID}&path={requestPath}&payload={jsonBody}&timestamp={timestamp}`
       - Hash with client secret using HMAC-SHA256
       - Encode as hex string
     - **Step 9.3:** Set headers:
       - `Content-Type`: `application/json`
       - `X-SF-RequestID`: Generated UUID v4
       - `X-SF-AccessToken`: Access token
       - `X-SF-Timestamp`: Current Unix timestamp
       - `X-SF-AppID`: Client ID
       - `X-SF-Signature`: Generated signature

10. **Send Request to Shopee**
    - Execute HTTP PUT request using `http.Client`
    - If request fails, return error

11. **Response Handling**
    - Read response body
    - If status code is 503 (Service Unavailable):
      - Read error response body
      - Return error with status code and message
    - Parse JSON response into `ShopeeResponse` struct
    - If parsing fails, return error

12. **State Update**
    - Update `merchantSecret.ShopeeMenuSyncState = "TRIGGER"`
    - Save to database
    - If save fails:
      - Log error
      - Return error response with code 500

13. **Response**
    - Return `Response`:
      - `code`: Shopee response code
      - `message`: Shopee response message

---

## 5. Order Submission Endpoint

**Path:** `/api/shopeefood/orders`  
**Method:** `POST`  
**Authentication:** `auth` (called by Shopee)  
**Handler:** `ShopeeOrderSubmission`

### Step-by-Step Flow:

1. **Request Reception**
   - Shopee sends POST request with order data
   - Request payload: `OrderRequest{ID, StoreID, Items, Amount, PaymentMethod, CreateTime, etc.}`

2. **Request Caching**
   - Create a new `Cache` record in database
   - Cache key: `"shopeefood_order_submission"`
   - Cache value: String representation of entire request
   - Log any errors during caching (non-blocking)

3. **Database Transaction Start**
   - Begin database transaction
   - Set up defer function to rollback on panic

4. **Merchant Secret Lookup**
   - Query `MerchantSecret` table where `shopee_store_id = req.StoreID`
   - If not found:
     - Log error
     - Rollback transaction
     - Return error response with code 400

5. **Item Conversion**
   - Call `generateCartItemsFromShopeeItems(req.Items)`:
     - **Step 5.1:** For each Shopee item:
       - **Step 5.1.1:** Process option groups:
         - Validate `OptionGroup.ExternalID` is not empty
         - Convert `ExternalID` string to UUID
         - Query `ModifierGroups` table to verify existence
         - For each option in group:
           - Validate `Option.ExternalID` is not empty
           - Convert `ExternalID` string to UUID
           - Extract price adjustment (divide by 100 if present)
           - Build `ModifierOption` with ID, name, price adjustment, quantity
         - Build `SelectedModifierGroup` with group ID, name, input type, options
       - **Step 5.1.2:** Process dish:
         - Validate `Dish.ExternalID` is not empty
         - Convert `ExternalID` string to UUID
       - **Step 5.1.3:** Build `CartItem`:
         - `ID`: Dish UUID
         - `Quantity`: Item quantity
         - `UnitPrice`: Unit price / 100 (convert cents to currency)
         - `SubTotal`: Subtotal / 100
         - `ItemNotes`: Item remark
         - `SelectedModifierGroups`: Generated modifier groups
     - **Step 5.2:** Return array of `CartItem`
   - If conversion fails:
     - Rollback transaction
     - Return error response with code 400

6. **Order Date Processing**
   - Convert `req.CreateTime` (Unix timestamp) to `time.Time`
   - If conversion results in zero time, use current time

7. **Order Request Construction**
   - Build `CreateOrderRequest`:
     - `OutletID`: From merchant secret
     - `OrderType`: `OrderTypeDelivery`
     - `OrderStatus`: `OrderStatusPending`
     - `OrderDate`: Processed order date
     - `InvoiceNumber`: `req.ID` (Shopee order ID)
     - `GrossTotal`: `req.Amount.TotalAmount / 100`
     - `NetTotal`: `req.Amount.TotalAmount / 100`
     - `ServiceCharge`: `req.Amount.MerchantServiceFee / 100`
     - `TaxCharge`: `req.Amount.TaxAmount / 100`
     - `TaxPercentage`: Calculated from tax amount / subtotal
     - `PaymentMethod`: `req.PaymentMethod`
     - `Notes`: `req.Remark`
     - `CartItems`: Converted cart items
     - `Platform`: `PlatformShopeeFood`
     - `PlatformOrderID`: `req.ID`
     - `PlatformState`: `"ACCEPTED"`
     - `CustomerName`: `nil` (Shopee doesn't provide)
     - `CustomerPhone`: `nil`
     - `CustomerAddress`: `nil`
     - `CustomerLatitude`: `nil`
     - `CustomerLongitude`: `nil`

8. **Order Creation**
   - Call `common_operations.CreateOrder(trx, &createOrderRequest)`
   - If creation fails:
     - Rollback transaction
     - Return error response with code 400

9. **Transaction Record Processing**
   - Query `Transaction` table where `order_id = order.ID`
   - Preload `Order` relationship
   - If not found:
     - Rollback transaction
     - Return error response with code 400

10. **Transaction Number Generation**
    - Generate transaction number using `payment.GenerateTransactionNumber(order.OrderNumber)`
    - Update transaction:
      - `TransactionNumber`: Generated number
      - `PaymentStatus`: `PaymentStatusCompleted`
    - Save transaction
    - If save fails:
      - Rollback transaction
      - Return error response with code 400

11. **Order Payment Status Update**
    - Update `order.PaymentStatus = PaymentStatusCompleted`
    - Save order
    - If save fails:
      - Rollback transaction
      - Return error response with code 400

12. **User Notification Preparation**
    - Query all `User` records where `outlet_id = merchantSecret.ID`
    - If query fails:
      - Rollback transaction
      - Return error response with code 400

13. **Notification Creation**
    - For each user:
      - Create `Notification` record:
        - `OutletID`: Merchant secret ID
        - `UserID`: User ID
        - `FCMDeviceToken`: User's FCM token
        - `Title`: "Hooray There is a new order"
        - `Body`: "Order #{orderNumber} received. View details in the app."
        - `NotificationType`: `ShopeeFoodNotification`
        - `IsRead`: `false`
      - Collect notification IDs and device tokens

14. **Transaction Commit**
    - Commit database transaction
    - If commit fails:
      - Rollback transaction
      - Return error response with code 400

15. **Firebase Notification**
    - After successful commit, send Firebase notifications:
      - Call `firebase.SendNotificationToMultipleDevices()`
      - Pass device tokens, title, body, notification IDs
      - Action URL: `"/streetfood/order"`
      - Notification type: `ShopeeFoodNotification`

16. **Response**
    - Return `Response`:
      - `code`: 0
      - `message`: "Order received and processed successfully"

---

## 6. Order Status Update Endpoint

**Path:** `/api/shopeefood/order/status`  
**Method:** `POST`  
**Authentication:** `auth` (called by Shopee)  
**Handler:** `ShopeePushOrderStatus`

### Step-by-Step Flow:

1. **Request Reception**
   - Shopee sends POST request with order status update
   - Request payload: `OrderStateRequest{ID, StoreID, Status, Message, PickupSeq}`
   - Log order ID and status

2. **Request Caching**
   - Create a new `Cache` record in database
   - Cache key: `"shopeefood_push_order_status"`
   - Cache value: String representation of entire request
   - Log any errors during caching (non-blocking)

3. **Request Validation**
   - Validate request struct using `govalidator.ValidateStruct()`
   - Ensure `ID` and `Status` are present
   - If validation fails:
     - Return error response with code 400

4. **Order Lookup**
   - Query `Order` table where `platform_order_id = req.ID`
   - Preload `OrderDetails` relationship
   - If not found:
     - Log error
     - Return error response with code 404 and message "Order not found"

5. **Platform State Update**
   - Update `order.PlatformState = req.Status`

6. **Status-Specific Processing**
   - **If status is CANCELED or CANCEL:**
     - Update `order.OrderStatus = OrderStatusCancelled`
     - **Step 6.1:** Prepare cancellation notification:
       - Title: "Order #{orderNumber} cancelled"
       - Body: "Order #{orderNumber} cancelled: {reason}. View details in the app."
     - **Step 6.2:** Query users for outlet
     - **Step 6.3:** Create notifications for each user
     - **Step 6.4:** Send Firebase notifications
   - **If status is DELIVERED or COMPLETED:**
     - Update `order.OrderStatus = OrderStatusCompleted`
     - Call `common_operations.ProcessSale(db, ctx, order.ID)` to process sale
   - **If status is PICKED_UP:**
     - Update `order.OrderStatus = OrderStatusOnTheWay`
   - **If status is REJECT:**
     - Update `order.OrderStatus = OrderStatusCancelled`
   - **If status is ACCEPTED:**
     - Update `order.OrderStatus = OrderStatusPreparing`

7. **Order Save**
   - Save updated order to database
   - If save fails:
     - Log error
     - Return error response with code 500 and message "Error updating order status"

8. **Response**
    - Log successful update
    - Return `Response`:
      - `code`: 0
      - `message`: "Order status updated successfully"

---

## 7. Get Order from Shopee Endpoint

**Path:** `/api/shopeefood/order/get-order/:order_id`  
**Method:** `GET`  
**Authentication:** `auth` (internal API)  
**Handler:** `GetOrderFromShopee`

### Step-by-Step Flow:

1. **Request Reception**
   - Internal system sends GET request with `order_id` as path parameter
   - Extract `order_id` from URL path

2. **Order Lookup**
   - Query `Order` table where `platform_order_id = order_id`
   - If not found:
     - Log error
     - Return error

3. **Merchant Secret Lookup**
   - Query `MerchantSecret` table where `outlet_id = order.OutletID`
   - If not found:
     - Log error
     - Return error

4. **Store ID Validation**
   - Check if `merchantSecret.ShopeeStoreID` is set
   - If nil:
     - Log error
     - Return error: "ShopeeStoreID is missing for this order's outlet"

5. **Business Configuration Lookup**
   - Query `BusinessConfiguration` table where `business_id = order.BusinessID`
   - If not found:
     - Log error
     - Return error: "business configuration not found"

6. **Credentials Validation**
   - Check if `ShopeeClientID` and `ShopeeClientSecret` are present
   - If missing:
     - Return error: "shopee client credentials are missing"

7. **Client Secret Decryption**
   - Decrypt `businessConfiguration.ShopeeClientSecret` using `common.DecryptText()`
   - If decryption fails:
     - Log error
     - Return error: "failed to decrypt Shopee client secret"

8. **Shopee API URL Construction**
   - Get base URL from `getShopeeBaseURL()`
   - Get `clientID` from business configuration
   - Construct full URL: `{baseURL}/api/adaptor/{clientID}/v1/orders/{order_id}?store_id={shopeeStoreID}`

9. **HTTP Request Creation**
   - Create GET request to Shopee API
   - No request body for GET request

10. **Request Header Preparation**
    - Call `headerPrepWithoutBody()` to prepare Shopee authentication headers:
      - **Step 10.1:** Get Shopee access token (same as Step 9.1 in Trigger Menu Update)
      - **Step 10.2:** Generate HMAC-SHA256 signature:
        - Extract query parameters from URL
        - Create signature string: `access_token={token}&app_id={clientID}&path={requestPath}&payload={queryParams}&timestamp={timestamp}`
        - Hash with client secret using HMAC-SHA256
        - Encode as hex string
      - **Step 10.3:** Set headers (same as Step 9.3 in Trigger Menu Update)

11. **Send Request to Shopee**
    - Execute HTTP GET request using `http.Client`
    - If request fails:
      - Log error
      - Return error: "failed to send request to ShopeeFood"

12. **Response Reading**
    - Read response body
    - If read fails:
      - Log error
      - Return error: "failed to read ShopeeFood response"

13. **Response Parsing**
    - Log raw response body
    - Unmarshal JSON response into `GetOrderResponse` struct
    - If parsing fails:
      - Log error
      - Return error: "failed to parse ShopeeFood response data"

14. **Response**
    - Return `GetOrderResponse`:
      - `code`: HTTP status code from Shopee
      - `message`: "success"
      - `data`: Parsed order data from Shopee

---

## Error Handling

### General Error Handling Patterns:
- All endpoints use structured error handling with detailed logging
- Database transaction rollbacks are in place for order submissions on any error
- Shopee error responses are parsed and included in output for observability
- Validation errors return appropriate HTTP status codes (400, 404, 500)
- Authentication failures return `Unauthenticated` error

### Specific Error Scenarios:
- **Missing Merchant Secret:** Returns 400/404 with descriptive error message
- **Invalid External IDs:** Returns 400 with detailed validation error
- **Database Transaction Failures:** Automatic rollback with error logging
- **Shopee API Failures:** Logged and returned with original error details
- **Decryption Failures:** Logged and returned with error message

---

## Authentication & Security

### Shopee-to-System Authentication:
- OAuth token endpoint validates partner credentials against system configuration
- JWT tokens are generated with 24-hour expiration
- All Shopee callbacks use `auth` middleware

### System-to-Shopee Authentication:
- Access tokens obtained via OAuth client credentials flow
- HMAC-SHA256 signatures generated for each request
- Signatures include: access_token, app_id, path, payload, timestamp
- Headers required: `X-SF-RequestID`, `X-SF-AccessToken`, `X-SF-Timestamp`, `X-SF-AppID`, `X-SF-Signature`

---

## Data Flow Diagrams

### Order Submission Flow:
```
Shopee → POST /api/shopeefood/orders
  → Cache Request
  → Lookup MerchantSecret
  → Convert Items to CartItems
  → Create Order (Transaction)
  → Update Transaction
  → Create Notifications
  → Commit Transaction
  → Send Firebase Notifications
  → Return Success Response
```

### Menu Sync Flow:
```
Internal System → GET /api/shopeefood/menu/trigger-update/:outletID
  → Lookup Outlet & Config
  → Prepare Shopee API Request
  → Get Access Token
  → Generate Signature
  → PUT to Shopee API
  → Update Sync State to "TRIGGER"
  → Shopee → GET /api/shopeefood/menu
  → Generate Catalog
  → Return Menu
  → Shopee → POST /api/shopeefood/menu/notification/result
  → Update Sync State to "SUCCESS" or "FAILED"
```

---

## Reference

### Implementation Files:
- [shopeefood/handlers.go](../../shopeefood/handlers.go) - Main endpoint handlers
- [shopeefood/order.go](../../shopeefood/order.go) - Order conversion logic
- [shopeefood/menu.go](../../shopeefood/menu.go) - Menu generation logic
- [shopeefood/client.go](../../shopeefood/client.go) - Shopee API client utilities
- [shopeefood/types.go](../../shopeefood/types.go) - Request/response type definitions
- [shopeefood/utils.go](../../shopeefood/utils.go) - Utility functions

### Related Operations:
- `common_operations.CreateOrder()` - Internal order creation
- `common_operations.ProcessSale()` - Sale processing for completed orders
- `common_operations.InsertNotification()` - Notification creation
- `firebase.SendNotificationToMultipleDevices()` - Firebase push notifications


