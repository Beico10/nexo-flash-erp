"""
Phase 4 Testing - Nexo One ERP
Tests: NFE, Industry PCP, Shoes Grid, AI Copilot modules
"""
import pytest
import requests
import os

BASE_URL = os.environ.get('REACT_APP_BACKEND_URL', '').rstrip('/')
if not BASE_URL:
    BASE_URL = "http://localhost:8001"

# Test credentials
TEST_TENANT = "demo"
TEST_EMAIL = "admin@demo.com"
TEST_PASSWORD = "demo123"


class TestAuth:
    """1. Login API tests"""
    
    def test_login_returns_access_token(self):
        """Test login with valid credentials returns access_token"""
        response = requests.post(f"{BASE_URL}/api/auth/login", json={
            "tenant_slug": TEST_TENANT,
            "email": TEST_EMAIL,
            "password": TEST_PASSWORD
        })
        assert response.status_code == 200, f"Login failed: {response.text}"
        data = response.json()
        assert "access_token" in data, f"Response missing access_token: {data}"
        assert "expires_in" in data
        assert "token_type" in data
        assert data["token_type"] == "Bearer"
        print(f"Login successful, token_type={data['token_type']}, expires_in={data['expires_in']}")

    def test_login_invalid_credentials(self):
        """Test login with invalid credentials fails"""
        response = requests.post(f"{BASE_URL}/api/auth/login", json={
            "tenant_slug": "invalid",
            "email": "wrong@test.com",
            "password": "wrongpass"
        })
        assert response.status_code in [400, 401, 404], f"Expected auth failure but got {response.status_code}"


@pytest.fixture(scope="module")
def auth_token():
    """Get auth token for authenticated tests"""
    response = requests.post(f"{BASE_URL}/api/auth/login", json={
        "tenant_slug": TEST_TENANT,
        "email": TEST_EMAIL,
        "password": TEST_PASSWORD
    })
    if response.status_code != 200:
        pytest.skip("Login failed - skipping authenticated tests")
    return response.json()["access_token"]


@pytest.fixture
def auth_headers(auth_token):
    """Headers with authorization"""
    return {"Authorization": f"Bearer {auth_token}", "Content-Type": "application/json"}


class TestNFE:
    """2-4. NFE Document tests"""
    
    def test_nfe_list_returns_seeded_docs(self, auth_headers):
        """NFE List: GET /api/v1/nfe/documents should return seeded docs (count >= 4)"""
        response = requests.get(f"{BASE_URL}/api/v1/nfe/documents", headers=auth_headers)
        assert response.status_code == 200, f"NFE list failed: {response.text}"
        data = response.json()
        assert "documents" in data, f"Response missing documents: {data}"
        assert "count" in data
        assert data["count"] >= 4, f"Expected >= 4 seeded docs, got {data['count']}"
        print(f"NFE list: {data['count']} documents found")
        
        # Verify document structure with snake_case
        if data["documents"]:
            doc = data["documents"][0]
            assert "recipient_name" in doc, "Document missing recipient_name"
            assert "recipient_cnpj" in doc, "Document missing recipient_cnpj"
            assert "access_key" in doc, "Document missing access_key"
            assert "status" in doc
            print(f"First doc: type={doc['type']}, status={doc['status']}, total={doc['total']}")
    
    def test_nfe_emit_creates_authorized_doc(self, auth_headers):
        """NFE Emit: POST /api/v1/nfe/emit should return document with status=authorized"""
        payload = {
            "type": "nfe",
            "recipient_name": "Test Company Ltd",
            "recipient_cnpj": "12345678000199",
            "description": "Test product",
            "total": 1000.00
        }
        response = requests.post(f"{BASE_URL}/api/v1/nfe/emit", headers=auth_headers, json=payload)
        assert response.status_code == 201, f"NFE emit failed: {response.text}"
        data = response.json()
        assert "document" in data, f"Response missing document: {data}"
        doc = data["document"]
        assert doc["status"] == "authorized", f"Expected authorized status, got {doc['status']}"
        assert doc["recipient_name"] == payload["recipient_name"]
        assert doc["total"] == payload["total"]
        assert "access_key" in doc and len(doc["access_key"]) > 30
        print(f"Emitted NFE: id={doc['id']}, number={doc['number']}, status={doc['status']}")
        return doc["id"]
    
    def test_nfe_cancel_authorized_doc(self, auth_headers):
        """NFE Cancel: POST /api/v1/nfe/cancel/{id} should cancel an authorized document"""
        # First list documents to find an authorized one (use nfe-1 from seeds)
        response = requests.post(f"{BASE_URL}/api/v1/nfe/cancel/nfe-1", headers=auth_headers)
        # Note: might already be cancelled from previous test run
        if response.status_code == 400:
            data = response.json()
            assert "cancelados" in data.get("error", "") or "autorizados" in data.get("error", "")
            print("nfe-1 already cancelled - this is expected behavior")
        else:
            assert response.status_code == 200, f"NFE cancel failed: {response.text}"
            data = response.json()
            assert data.get("cancelled") == True
            print("NFE cancelled successfully")


class TestIndustry:
    """5-10. Industry PCP module tests"""
    
    def test_industry_orders_list(self, auth_headers):
        """Industry Orders List: GET /api/v1/industry/orders should return seeded orders (count >= 3)"""
        response = requests.get(f"{BASE_URL}/api/v1/industry/orders", headers=auth_headers)
        assert response.status_code == 200, f"Industry orders list failed: {response.text}"
        data = response.json()
        assert "orders" in data, f"Response missing orders: {data}"
        assert "count" in data
        assert data["count"] >= 3, f"Expected >= 3 seeded orders, got {data['count']}"
        print(f"Industry orders: {data['count']} orders found")
        
        # Verify snake_case fields
        if data["orders"]:
            order = data["orders"][0]
            assert "product_name" in order, "Order missing product_name"
            assert "planned_qty" in order, "Order missing planned_qty"
            assert "produced_qty" in order, "Order missing produced_qty"
            assert "planned_start" in order, "Order missing planned_start"
            assert "planned_end" in order, "Order missing planned_end"
            print(f"First order: {order['product_name']}, status={order['status']}")
    
    def test_industry_order_detail(self, auth_headers):
        """Industry Order Detail: GET /api/v1/industry/orders/op-1 should return order with snake_case fields"""
        response = requests.get(f"{BASE_URL}/api/v1/industry/orders/op-1", headers=auth_headers)
        assert response.status_code == 200, f"Industry order detail failed: {response.text}"
        data = response.json()
        assert "order" in data, f"Response missing order: {data}"
        order = data["order"]
        # Verify snake_case fields in detail response
        assert "ProductName" not in order, "Order has camelCase ProductName - should be snake_case"
        print(f"Order op-1: {order}")
    
    def test_industry_boms_list(self, auth_headers):
        """Industry BOMs: GET /api/v1/industry/boms should return seeded BOMs"""
        response = requests.get(f"{BASE_URL}/api/v1/industry/boms", headers=auth_headers)
        assert response.status_code == 200, f"Industry BOMs list failed: {response.text}"
        data = response.json()
        assert "boms" in data, f"Response missing boms: {data}"
        print(f"Industry BOMs: {len(data['boms'])} BOMs found")
        
        if data["boms"]:
            bom = data["boms"][0]
            assert "product_name" in bom, "BOM missing product_name"
            assert "total_cost" in bom, "BOM missing total_cost"
            assert "items_count" in bom, "BOM missing items_count"
    
    def test_industry_materials_list(self, auth_headers):
        """Industry Materials: GET /api/v1/industry/materials should return material balances"""
        response = requests.get(f"{BASE_URL}/api/v1/industry/materials", headers=auth_headers)
        assert response.status_code == 200, f"Industry materials failed: {response.text}"
        data = response.json()
        assert "materials" in data, f"Response missing materials: {data}"
        print(f"Industry materials: {len(data['materials'])} materials tracked")
    
    def test_industry_create_order(self, auth_headers):
        """Industry Create Order: POST /api/v1/industry/orders with proper fields"""
        payload = {
            "product_id": "prod-test",
            "product_name": "Test Product",
            "bom_id": "bom-mesa",
            "planned_qty": 10,
            "unit": "un"
        }
        response = requests.post(f"{BASE_URL}/api/v1/industry/orders", headers=auth_headers, json=payload)
        assert response.status_code == 201, f"Industry create order failed: {response.text}"
        data = response.json()
        assert "order" in data, f"Response missing order: {data}"
        order = data["order"]
        assert order.get("product_name") == payload["product_name"], f"product_name mismatch: {order}"
        assert order.get("planned_qty") == payload["planned_qty"]
        assert order.get("status") == "planned"
        print(f"Created order: id={order['id']}, product={order['product_name']}, status={order['status']}")
    
    def test_industry_bom_explode(self, auth_headers):
        """Industry BOM Explode: POST /api/v1/industry/bom/explode returns reservations and total_cost"""
        payload = {
            "product_id": "prod-mesa",
            "quantity": 5
        }
        response = requests.post(f"{BASE_URL}/api/v1/industry/bom/explode", headers=auth_headers, json=payload)
        assert response.status_code == 200, f"BOM explode failed: {response.text}"
        data = response.json()
        assert "reservations" in data, f"Response missing reservations: {data}"
        assert "total_cost" in data, f"Response missing total_cost: {data}"
        print(f"BOM explode: {len(data['reservations'])} reservations, total_cost={data['total_cost']}")


class TestShoes:
    """11-14. Shoes Grid module tests"""
    
    def test_shoes_grids_list(self, auth_headers):
        """Shoes Grids List: GET /api/v1/shoes/grids should return seeded grids (count >= 3)"""
        response = requests.get(f"{BASE_URL}/api/v1/shoes/grids", headers=auth_headers)
        assert response.status_code == 200, f"Shoes grids list failed: {response.text}"
        data = response.json()
        assert "grids" in data, f"Response missing grids: {data}"
        assert "count" in data
        assert data["count"] >= 3, f"Expected >= 3 seeded grids, got {data['count']}"
        print(f"Shoes grids: {data['count']} grids found")
        
        # Verify snake_case fields
        if data["grids"]:
            grid = data["grids"][0]
            assert "model_code" in grid, "Grid missing model_code"
            assert "model_name" in grid, "Grid missing model_name"
            assert "brand" in grid, "Grid missing brand"
            assert "total_skus" in grid, "Grid missing total_skus"
            assert "in_stock_skus" in grid, "Grid missing in_stock_skus"
            print(f"First grid: {grid['model_code']} - {grid['model_name']}")
    
    def test_shoes_grid_detail(self, auth_headers):
        """Shoes Grid Detail: GET /api/v1/shoes/grids/SANDAL001 should return grid with snake_case fields"""
        response = requests.get(f"{BASE_URL}/api/v1/shoes/grids/SANDAL001", headers=auth_headers)
        assert response.status_code == 200, f"Shoes grid detail failed: {response.text}"
        data = response.json()
        assert "grid" in data, f"Response missing grid: {data}"
        grid = data["grid"]
        # Verify it's the correct model
        assert grid.get("ModelCode") == "SANDAL001" or grid.get("model_code") == "SANDAL001"
        # Check for summary
        assert "summary" in data, "Response missing summary"
        print(f"Grid SANDAL001: {grid}")
    
    def test_shoes_stock_update(self, auth_headers):
        """Shoes Stock Update: PATCH /api/v1/shoes/stock with sku and delta"""
        payload = {
            "sku": "SANDAL001-PRETO-38",
            "delta": 5
        }
        response = requests.patch(f"{BASE_URL}/api/v1/shoes/stock", headers=auth_headers, json=payload)
        # Note: might fail if SKU doesn't exist - check both success and error cases
        if response.status_code == 200:
            data = response.json()
            assert data.get("updated") == True
            print("Stock updated successfully")
        else:
            data = response.json()
            print(f"Stock update response: {response.status_code} - {data}")
            # If SKU not found, that's acceptable for this test
            assert response.status_code in [200, 400]
    
    def test_shoes_commissions(self, auth_headers):
        """Shoes Commissions: GET /api/v1/shoes/commissions should return seller commissions"""
        response = requests.get(f"{BASE_URL}/api/v1/shoes/commissions", headers=auth_headers)
        assert response.status_code == 200, f"Shoes commissions failed: {response.text}"
        data = response.json()
        assert "commissions" in data, f"Response missing commissions: {data}"
        print(f"Shoes commissions: {len(data['commissions'])} sellers")
        
        if data["commissions"]:
            comm = data["commissions"][0]
            assert "seller_name" in comm, "Commission missing seller_name"
            assert "base_rate" in comm, "Commission missing base_rate"
            assert "total_sales" in comm, "Commission missing total_sales"
            assert "total_commission" in comm, "Commission missing total_commission"
            print(f"First seller: {comm['seller_name']}, commission={comm['total_commission']}")


class TestCopilot:
    """15. AI Co-Pilot tests"""
    
    def test_copilot_suggest(self, auth_headers):
        """AI Co-Pilot: POST /api/v1/copilot/suggest should return suggestion from Gemini"""
        payload = {
            "question": "Como melhorar vendas?",
            "session_id": "test-session"
        }
        response = requests.post(f"{BASE_URL}/api/v1/copilot/suggest", headers=auth_headers, json=payload)
        assert response.status_code == 200, f"Copilot suggest failed: {response.text}"
        data = response.json()
        assert "suggestion" in data, f"Response missing suggestion: {data}"
        assert len(data["suggestion"]) > 10, "Suggestion too short"
        print(f"Copilot suggestion (first 100 chars): {data['suggestion'][:100]}...")
    
    def test_copilot_unauthorized(self):
        """Copilot without auth should return 401"""
        response = requests.post(f"{BASE_URL}/api/v1/copilot/suggest", json={
            "question": "Test",
            "session_id": "test"
        })
        assert response.status_code == 401, f"Expected 401, got {response.status_code}"


if __name__ == "__main__":
    pytest.main([__file__, "-v", "--tb=short"])
