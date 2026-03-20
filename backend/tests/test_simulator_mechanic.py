"""
Nexo One ERP - Iteration 2 Tests
Tests: Public Tax Simulator endpoints (no auth) and Mechanic OS with real backend API
"""
import pytest
import requests
import os

BASE_URL = os.environ.get('REACT_APP_BACKEND_URL', 'https://erp-fiscal-demo.preview.emergentagent.com').rstrip('/')

# Test credentials
TEST_TENANT = "demo"
TEST_EMAIL = "admin@demo.com"
TEST_PASSWORD = "demo123"


@pytest.fixture(scope="session")
def api_client():
    """Shared requests session without auth"""
    session = requests.Session()
    session.headers.update({"Content-Type": "application/json"})
    return session


@pytest.fixture(scope="session")
def auth_token(api_client):
    """Get authentication token"""
    response = api_client.post(f"{BASE_URL}/api/auth/login", json={
        "tenant_slug": TEST_TENANT,
        "email": TEST_EMAIL,
        "password": TEST_PASSWORD
    })
    if response.status_code == 200:
        return response.json().get("access_token")
    pytest.fail(f"Authentication failed: {response.text}")


@pytest.fixture(scope="session")
def authenticated_client(api_client, auth_token):
    """Session with auth header"""
    auth_session = requests.Session()
    auth_session.headers.update({
        "Content-Type": "application/json",
        "Authorization": f"Bearer {auth_token}"
    })
    return auth_session


# ============================================================================
# PUBLIC TAX SIMULATOR TESTS (NO AUTH REQUIRED)
# ============================================================================

class TestPublicNCMList:
    """GET /api/v1/tax/ncm-list - Public endpoint without authentication"""

    def test_ncm_list_no_auth_returns_200(self, api_client):
        """NCM list endpoint should work without authentication"""
        response = api_client.get(f"{BASE_URL}/api/v1/tax/ncm-list")
        assert response.status_code == 200, f"Expected 200, got {response.status_code}"
        print(f"✓ NCM list accessible without auth: status={response.status_code}")

    def test_ncm_list_returns_data_array(self, api_client):
        """NCM list should return data array with items"""
        response = api_client.get(f"{BASE_URL}/api/v1/tax/ncm-list")
        data = response.json()
        assert "data" in data, "Response missing 'data' field"
        assert "total" in data, "Response missing 'total' field"
        assert isinstance(data["data"], list), "data should be a list"
        assert len(data["data"]) > 0, "data should not be empty"
        print(f"✓ NCM list structure: total={data['total']}, items={len(data['data'])}")

    def test_ncm_list_item_structure(self, api_client):
        """Each NCM item should have required fields"""
        response = api_client.get(f"{BASE_URL}/api/v1/tax/ncm-list")
        data = response.json()
        item = data["data"][0]
        required_fields = ["ncm_code", "description", "ibs_rate", "cbs_rate", "selective_rate", "is_basket"]
        for field in required_fields:
            assert field in item, f"Missing required field: {field}"
        print(f"✓ NCM item structure valid: {item['ncm_code']} - {item['description']}")


class TestPublicSimulate:
    """POST /api/v1/tax/simulate - Public tax simulation without authentication"""

    def test_simulate_no_auth_returns_200(self, api_client):
        """Simulate endpoint should work without authentication"""
        response = api_client.post(f"{BASE_URL}/api/v1/tax/simulate", json={
            "ncm_code": "84715010",
            "base_value": 1000,
            "operation": "debit_exit"
        })
        assert response.status_code == 200, f"Expected 200, got {response.status_code}: {response.text}"
        print(f"✓ Simulate accessible without auth: status={response.status_code}")

    def test_simulate_cesta_basica_zero_tax(self, api_client):
        """NCM 19052000 (pao frances - cesta basica) should return total_tax: 0 and is_basket_item: true"""
        response = api_client.post(f"{BASE_URL}/api/v1/tax/simulate", json={
            "ncm_code": "19052000",
            "base_value": 100,
            "operation": "debit_exit"
        })
        assert response.status_code == 200
        data = response.json()
        assert data["total_tax"] == 0, f"Expected total_tax=0 for cesta basica, got {data['total_tax']}"
        assert data["is_basket_item"] == True, f"Expected is_basket_item=True, got {data['is_basket_item']}"
        assert "Cesta Basica" in data["legal_basis"], f"Legal basis should mention Cesta Basica: {data['legal_basis']}"
        print(f"✓ Cesta Basica NCM 19052000: total_tax={data['total_tax']}, is_basket_item={data['is_basket_item']}")

    def test_simulate_standard_ncm_84715010(self, api_client):
        """NCM 84715010 (notebook) with base_value 5000 should return ibs_amount: 462.5"""
        response = api_client.post(f"{BASE_URL}/api/v1/tax/simulate", json={
            "ncm_code": "84715010",
            "base_value": 5000,
            "operation": "debit_exit"
        })
        assert response.status_code == 200
        data = response.json()
        assert data["ibs_amount"] == 462.5, f"Expected ibs_amount=462.5, got {data['ibs_amount']}"
        assert data["cbs_amount"] == 187.5, f"Expected cbs_amount=187.5, got {data['cbs_amount']}"
        assert data["total_tax"] == 650, f"Expected total_tax=650, got {data['total_tax']}"
        print(f"✓ Standard NCM 84715010: ibs_amount={data['ibs_amount']}, cbs_amount={data['cbs_amount']}, total_tax={data['total_tax']}")

    def test_simulate_cerveja_selective_tax(self, api_client):
        """NCM 22030000 (cerveja) should return selective_amount > 0"""
        response = api_client.post(f"{BASE_URL}/api/v1/tax/simulate", json={
            "ncm_code": "22030000",
            "base_value": 1000,
            "operation": "debit_exit"
        })
        assert response.status_code == 200
        data = response.json()
        assert data["selective_amount"] > 0, f"Expected selective_amount > 0, got {data['selective_amount']}"
        assert data["selective_rate"] == 0.1, f"Expected selective_rate=0.1, got {data['selective_rate']}"
        print(f"✓ Cerveja NCM 22030000: selective_amount={data['selective_amount']}, selective_rate={data['selective_rate']}")

    def test_simulate_invalid_ncm_returns_error(self, api_client):
        """Invalid NCM should return 400 with error message"""
        response = api_client.post(f"{BASE_URL}/api/v1/tax/simulate", json={
            "ncm_code": "99999999",
            "base_value": 100,
            "operation": "debit_exit"
        })
        assert response.status_code == 400, f"Expected 400 for invalid NCM, got {response.status_code}"
        data = response.json()
        assert "error" in data, "Response should contain 'error' field"
        print(f"✓ Invalid NCM rejected: error={data['error']}")

    def test_simulate_response_structure(self, api_client):
        """Simulate response should have all required fields for frontend"""
        response = api_client.post(f"{BASE_URL}/api/v1/tax/simulate", json={
            "ncm_code": "84715010",
            "base_value": 1000,
            "operation": "debit_exit"
        })
        data = response.json()
        required_fields = [
            "ncm_code", "base_value", "ibs_rate", "cbs_rate", "selective_rate",
            "ibs_amount", "cbs_amount", "selective_amount", "total_tax",
            "cashback_amount", "is_basket_item", "legal_basis", "approval_status"
        ]
        for field in required_fields:
            assert field in data, f"Missing required field: {field}"
        print(f"✓ Response structure complete with all {len(required_fields)} required fields")


# ============================================================================
# MECHANIC OS TESTS (AUTHENTICATED - Real Backend)
# ============================================================================

class TestMechanicOS:
    """Mechanic Service Order CRUD - Connected to real backend API"""

    def test_list_os_authenticated(self, authenticated_client):
        """GET /api/v1/mechanic/os should return OS list with snake_case JSON"""
        response = authenticated_client.get(f"{BASE_URL}/api/v1/mechanic/os")
        assert response.status_code == 200
        data = response.json()
        assert "data" in data
        assert "total" in data
        # data can be None if no OS exists, or a list
        os_list = data["data"] or []
        assert isinstance(os_list, list)
        print(f"✓ List OS: total={data['total']}, count={len(os_list)}")

    def test_create_os_and_verify_in_list(self, authenticated_client):
        """POST /api/v1/mechanic/os should create OS and appear in list"""
        # Create OS with unique plate
        import time
        unique_plate = f"TST{int(time.time()) % 10000:04d}"
        
        create_response = authenticated_client.post(f"{BASE_URL}/api/v1/mechanic/os", json={
            "vehicle_plate": unique_plate,
            "vehicle_km": 55000,
            "vehicle_model": "Test Model",
            "vehicle_year": 2023,
            "customer_id": "TEST_customer_create",
            "customer_phone": "11988887777",
            "complaint": "Test complaint for verification"
        })
        assert create_response.status_code == 201, f"Create failed: {create_response.text}"
        created = create_response.json()
        
        # Verify response structure (snake_case)
        assert "id" in created, "Response should have 'id' field (snake_case)"
        assert "number" in created, "Response should have 'number' field"
        assert created["vehicle_plate"] == unique_plate
        assert created["status"] == "open"
        
        os_id = created["id"]
        print(f"✓ Created OS: id={os_id}, number={created['number']}, plate={created['vehicle_plate']}")

        # Verify OS appears in list
        list_response = authenticated_client.get(f"{BASE_URL}/api/v1/mechanic/os")
        assert list_response.status_code == 200
        list_data = list_response.json()
        os_ids = [os["id"] for os in list_data["data"]]
        assert os_id in os_ids, f"Created OS {os_id} not found in list"
        print(f"✓ Created OS appears in list: verified")

    def test_os_response_snake_case(self, authenticated_client):
        """OS response should use snake_case (not PascalCase) for frontend compatibility"""
        response = authenticated_client.get(f"{BASE_URL}/api/v1/mechanic/os")
        data = response.json()
        
        if len(data["data"]) > 0:
            os_item = data["data"][0]
            # Check for snake_case fields (new format)
            snake_case_fields = ["id", "tenant_id", "number", "vehicle_plate", "vehicle_km", 
                                 "vehicle_model", "vehicle_year", "customer_id", "customer_phone",
                                 "status", "complaint", "created_at", "updated_at"]
            for field in snake_case_fields:
                assert field in os_item, f"Missing snake_case field: {field}"
            
            # Ensure NO PascalCase (old format)
            pascal_case_fields = ["ID", "TenantID", "VehiclePlate", "Status"]
            for field in pascal_case_fields:
                assert field not in os_item, f"Found PascalCase field {field} - should be snake_case"
            
            print(f"✓ OS response uses snake_case format: {list(os_item.keys())[:5]}...")
        else:
            pytest.skip("No OS items to verify field format")


# ============================================================================
# AUTH PROTECTED ENDPOINT VERIFICATION
# ============================================================================

class TestAuthProtection:
    """Verify auth-protected vs public endpoints"""

    def test_tax_calculate_requires_auth(self, api_client):
        """POST /api/v1/tax/calculate (not simulate) should require auth"""
        response = api_client.post(f"{BASE_URL}/api/v1/tax/calculate", json={
            "ncm_code": "84715010",
            "base_value": 1000,
            "operation": "debit_exit"
        })
        assert response.status_code == 401, f"Expected 401 for protected endpoint, got {response.status_code}"
        print(f"✓ /api/v1/tax/calculate requires auth: status={response.status_code}")

    def test_mechanic_os_requires_auth(self, api_client):
        """GET /api/v1/mechanic/os should require auth"""
        response = api_client.get(f"{BASE_URL}/api/v1/mechanic/os")
        assert response.status_code == 401, f"Expected 401 for protected endpoint, got {response.status_code}"
        print(f"✓ /api/v1/mechanic/os requires auth: status={response.status_code}")

    def test_public_simulate_no_auth(self, api_client):
        """POST /api/v1/tax/simulate should NOT require auth"""
        response = api_client.post(f"{BASE_URL}/api/v1/tax/simulate", json={
            "ncm_code": "84715010",
            "base_value": 1000,
            "operation": "debit_exit"
        })
        assert response.status_code == 200, f"Simulate should be public, got {response.status_code}"
        print(f"✓ /api/v1/tax/simulate is public: status={response.status_code}")

    def test_public_ncm_list_no_auth(self, api_client):
        """GET /api/v1/tax/ncm-list should NOT require auth"""
        response = api_client.get(f"{BASE_URL}/api/v1/tax/ncm-list")
        assert response.status_code == 200, f"NCM list should be public, got {response.status_code}"
        print(f"✓ /api/v1/tax/ncm-list is public: status={response.status_code}")


# ============================================================================
# RUN TESTS
# ============================================================================

if __name__ == "__main__":
    pytest.main([__file__, "-v", "--tb=short"])
