# **Products API**

## **Introduction**

The **Products API** provides a structured way for clients to manage "Products" within the system. It offers multiple endpoints that allow users to perform various operations related to **stock items**, **store products**, and **e-commerce platform integrations**.

### **Key Features**

- ✅ **Stock Item Management**: SQL account users can **POST** stock items for a specific company.
- ✅ **E-commerce Integrations**: Sellers can **retrieve unmapped products** from linked e-commerce stores.
- ✅ **Product Mapping**: Sellers can **map store products** to stock items and view/remove mapped products.
- ✅ **Scalable Architecture**: Developed using the **handlers, services, and repositories** pattern for maintainability.

## Notes

> - 🔹 **Access tokens** of e-commerce stores are **stored in the database** to optimize mapped/unmapped product retrieval.
> - 🔹 Uses **`int64`** for all **IDs** in repositories and models to ensure consistency.
> - 🔹 **Lazada SDK** is used to streamline API interactions with Lazada.
> - 🔹 The system follows the **handlers, services, and repositories** structure to ensure maintainability and scalability.

## **📌 Postman Collection as Sample for HTTP Request**

Click the button below to view sample HTTP Request

[![Run in Postman](https://run.pstmn.io/button.svg)](https://documenter.getpostman.com/view/39111263/2sAYX9nfUs)

## **📌 API Endpoints**

### **🛒 Stock Items Management**

1️⃣ **Retrieve Stock Items by Company**

- **`GET /products/stock-item/:company_id`**
- Retrieves all stock items associated with the given company.

2️⃣ **Post Stock Items from SQL Account**

- **`POST /products/stock-item/:company_id`**
- Inserts or updates stock items based on `company_id` and `stock_code`.
- **Logic:**
  - If a stock item **already exists**, it is updated instead of inserted.
  - Stock items **not in the latest request** are automatically deleted.
- **Note:** The combination of `company_id` and `stock_code` **must be unique**.

---

### **🏪 Store Products Management**

3️⃣ **Retrieve Store Products by Company**

- **`GET /products/store-products/:company_id`**
- Fetches store products for a specific company.

4️⃣ **Map Store Products to Stock Items**

- **`POST /products/store-products`**
- Maps e-commerce store products to stock items.
- **Required fields:**
  - `stock_item_id` → Specifies which stock item the product belongs to.
  - `store_id` → Specifies which store the product is linked to.

---

### **🔗 Mapped Products Management**

5️⃣ **Retrieve Mapped Products from E-commerce Platforms**

- **`GET /products/mapped-products/:company_id`**
- Fetches **all mapped products** for a company from e-commerce platforms.

6️⃣ **Unmap a Single Store Product**

- **`DELETE /products/mapped-product`**
- Deletes/unmaps **one** store product using:
  - `store_id`
  - `sku`

7️⃣ **Unmap Multiple Store Products**

- **`DELETE /products/mapped-products`**
- Deletes/unmaps **multiple** store products using:
  - `store_id`
  - List of **SKUs**

---

### **🚀 Unmapped Products Management**

8️⃣ **Fetch Unmapped Products from E-commerce Platforms**

- **`GET /products/unmapped-products/:company_id`**
- Retrieves **all unmapped products** from **linked e-commerce platforms**.
- Uses **stored access tokens** to fetch data from linked stores.

---

## **🔧 TODO (Upcoming Features)**

- ✅ **Expand mapped & unmapped product retrieval**
  - Add support for **Shopee** and **TikTok** alongside Lazada. The services is defined, can using the logic directly if no better implementation idea.
- ✅ **Optimize unmapped product retrieval**
  - Fetch data from **all linked platforms** using stored access tokens.
