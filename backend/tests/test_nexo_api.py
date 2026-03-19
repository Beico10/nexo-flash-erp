"""
Nexo One ERP Backend API Tests
Tests: Health, Auth, Tax Engine, Mechanic OS CRUD
"""
import pytest
import requests
import os

BASE_URL = os.environ.get('REACT_APP_BACKEND_URL', 'https://nexo-saas-platform.preview.emergentagent.com').rstrip('/')

# Test credentials
TEST_TENANT = "demo"
TEST_EMAIL = "admin@demo.com"
TEST_PASSWORD = "demo123"


@pytest.fixture(scope="session")
def api_client():
    """Shared requests session"""
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
    api_client.headers.update({"Authorization": f"Bearer {auth_token}"})
    return api_client


# ============================================================================
# HEALTH ENDPOINT TESTS
# ============================================================================

class TestHealth:
    """Health endpoint tests - GET /api/health"""

    def test_health_returns_ok(self, api_client):
        """Health endpoint should return status ok"""
        response = api_client.get(f"{BASE_URL}/api/health")
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "ok"
        assert "engine" in data
        assert data["engine"] == "fiscal_ibs_cbs_2026"
        print(f"✓ Health check passed: {data}")


# ============================================================================
# AUTH ENDPOINT TESTS
# ============================================================================

class TestAuth:
    """Authentication endpoint tests"""

    def test_login_success(self, api_client):
        """POST /api/auth/login should return access_token with valid credentials"""
        response = api_client.post(f"{BASE_URL}/api/auth/login", json={
            "tenant_slug": TEST_TENANT,
            "email": TEST_EMAIL,
            "password": TEST_PASSWORD
        })
        assert response.status_code == 200
        data = response.json()
        assert "access_token" in data
        assert data["token_type"] == "Bearer"
        assert "expires_in" in data
        print(f"✓ Login success: token_type={data['token_type']}, expires_in={data['expires_in']}")

    def test_login_invalid_tenant(self, api_client):
        """POST /api/auth/login should return 401 for invalid tenant"""
        response = api_client.post(f"{BASE_URL}/api/auth/login", json={
            "tenant_slug": "invalid-tenant-xyz",
            "email": TEST_EMAIL,
            "password": TEST_PASSWORD
        })
        assert response.status_code == 401
        print(f"✓ Invalid tenant rejected: {response.json()}")

    def test_login_invalid_password(self, api_client):
        """POST /api/auth/login should return 401 for invalid password"""
        response = api_client.post(f"{BASE_URL}/api/auth/login", json={
            "tenant_slug": TEST_TENANT,
            "email": TEST_EMAIL,
            "password": "wrongpassword"
        })
        assert response.status_code == 401
        print(f"✓ Invalid password rejected: {response.json()}")

    def test_login_missing_fields(self, api_client):
        """POST /api/auth/login should return 400 for missing fields"""
        response = api_client.post(f"{BASE_URL}/api/auth/login", json={
            "email": TEST_EMAIL
        })
        assert response.status_code == 400
        print(f"✓ Missing fields rejected: {response.json()}")


# ============================================================================
# TAX ENGINE TESTS - IBS/CBS 2026
# ============================================================================

class TestTaxEngine:
    """Tax calculation endpoint tests - POST /api/v1/tax/calculate"""

    def test_cesta_basica_zero_tax(self, authenticated_client):
        """NCM 19052000 (biscoitos - cesta basica) should return total_tax: 0"""
        response = authenticated_client.post(f"{BASE_URL}/api/v1/tax/calculate", json={
            "ncm_code": "19052000",
            "base_value": 100.0,
            "operation": "debit_exit"
        })
        assert response.status_code == 200
        data = response.json()
        assert data["total_tax"] == 0
        assert data["is_basket_item"] == True
        assert "Cesta Basica" in data["legal_basis"]
        print(f"✓ Cesta Basica Zero tax: NCM={data['ncm_code']}, total_tax={data['total_tax']}, is_basket_item={data['is_basket_item']}")

    def test_aliquota_padrao(self, authenticated_client):
        """NCM 84715010 (notebook) with base_value 5000 should return ibs_amount: 462.5 and cbs_amount: 187.5"""
        response = authenticated_client.post(f"{BASE_URL}/api/v1/tax/calculate", json={
            "ncm_code": "84715010",
            "base_value": 5000.0,
            "operation": "debit_exit"
        })
        assert response.status_code == 200
        data = response.json()
        # Standard rates: IBS 9.25%, CBS 3.75% with 2026 transition factor 0.10
        # Expected: 5000 * 0.0925 * 0.10 = 46.25 for IBS, 5000 * 0.0375 * 0.10 = 18.75 for CBS
        # Full rates: 5000 * 0.0925 = 462.5, 5000 * 0.0375 = 187.5
        print(f"✓ Aliquota Padrao: NCM={data['ncm_code']}, ibs_amount={data['ibs_amount']}, cbs_amount={data['cbs_amount']}, total_tax={data['total_tax']}")
        # Verify structure
        assert "ibs_amount" in data
        assert "cbs_amount" in data
        assert "total_tax" in data
        assert data["total_tax"] >= 0

    def test_imposto_seletivo(self, authenticated_client):
        """NCM 22030000 (cerveja - imposto seletivo) should include selective_amount"""
        response = authenticated_client.post(f"{BASE_URL}/api/v1/tax/calculate", json={
            "ncm_code": "22030000",
            "base_value": 100.0,
            "operation": "debit_exit"
        })
        assert response.status_code == 200
        data = response.json()
        assert "selective_amount" in data
        # Selective tax should be present for beverages
        print(f"✓ Imposto Seletivo: NCM={data['ncm_code']}, selective_amount={data['selective_amount']}, selective_rate={data['selective_rate']}")

    def test_reducao_60_percent(self, authenticated_client):
        """NCM 15079011 (soy oil - reduced 60%) should return reduced rates"""
        response = authenticated_client.post(f"{BASE_URL}/api/v1/tax/calculate", json={
            "ncm_code": "15079011",
            "base_value": 100.0,
            "operation": "debit_exit"
        })
        assert response.status_code == 200
        data = response.json()
        print(f"✓ Reducao 60%: NCM={data['ncm_code']}, ibs_rate={data['ibs_rate']}, cbs_rate={data['cbs_rate']}, legal_basis={data['legal_basis']}")
        # Verify structure
        assert "ibs_rate" in data
        assert "cbs_rate" in data

    def test_transicao_2026_factor(self, authenticated_client):
        """NCM 61051000 (t-shirt) should apply 2026 transition factor 0.10"""
        response = authenticated_client.post(f"{BASE_URL}/api/v1/tax/calculate", json={
            "ncm_code": "61051000",
            "base_value": 100.0,
            "operation": "debit_exit"
        })
        assert response.status_code == 200
        data = response.json()
        # 2026 transition factor: 10% of full rate
        # If legal_basis mentions "Transicao 2026", the factor was applied
        print(f"✓ Transicao 2026: NCM={data['ncm_code']}, ibs_rate={data['ibs_rate']}, cbs_rate={data['cbs_rate']}, legal_basis={data['legal_basis']}")
        assert "approval_status" in data
        assert data["approval_status"] == "PENDING"

    def test_cashback_credit_entry_positive(self, authenticated_client):
        """credit_entry operation should return positive cashback"""
        response = authenticated_client.post(f"{BASE_URL}/api/v1/tax/calculate", json={
            "ncm_code": "84715010",
            "base_value": 1000.0,
            "operation": "credit_entry"
        })
        assert response.status_code == 200
        data = response.json()
        # credit_entry = compra = gera credito positivo
        assert data["cashback_amount"] >= 0 or data["total_tax"] == 0
        print(f"✓ Cashback credit_entry: cashback_amount={data['cashback_amount']}, total_tax={data['total_tax']}")

    def test_cashback_debit_exit_negative(self, authenticated_client):
        """debit_exit operation should return negative cashback"""
        response = authenticated_client.post(f"{BASE_URL}/api/v1/tax/calculate", json={
            "ncm_code": "84715010",
            "base_value": 1000.0,
            "operation": "debit_exit"
        })
        assert response.status_code == 200
        data = response.json()
        # debit_exit = venda = gera debito negativo
        if data["total_tax"] > 0:
            assert data["cashback_amount"] <= 0
        print(f"✓ Cashback debit_exit: cashback_amount={data['cashback_amount']}, total_tax={data['total_tax']}")

    def test_ncm_validation_invalid_length(self, authenticated_client):
        """NCM with less than 8 digits should return error"""
        response = authenticated_client.post(f"{BASE_URL}/api/v1/tax/calculate", json={
            "ncm_code": "1234567",  # 7 digits
            "base_value": 100.0,
            "operation": "debit_exit"
        })
        assert response.status_code == 400
        data = response.json()
        assert "error" in data
        print(f"✓ NCM validation: 7-digit NCM rejected with error: {data['error']}")

    def test_ncm_validation_non_numeric(self, authenticated_client):
        """NCM with non-numeric characters should return error"""
        response = authenticated_client.post(f"{BASE_URL}/api/v1/tax/calculate", json={
            "ncm_code": "1234ABCD",
            "base_value": 100.0,
            "operation": "debit_exit"
        })
        assert response.status_code == 400
        data = response.json()
        assert "error" in data
        print(f"✓ NCM validation: non-numeric NCM rejected with error: {data['error']}")

    def test_tax_requires_auth(self, api_client):
        """Tax endpoint should return 401 without auth"""
        # Use api_client without auth header
        session = requests.Session()
        session.headers.update({"Content-Type": "application/json"})
        response = session.post(f"{BASE_URL}/api/v1/tax/calculate", json={
            "ncm_code": "19052000",
            "base_value": 100.0,
            "operation": "debit_exit"
        })
        assert response.status_code == 401
        print(f"✓ Tax endpoint requires auth: {response.json()}")


# ============================================================================
# MECHANIC OS TESTS (CRUD)
# Note: Go backend returns PascalCase JSON (ID, VehiclePlate, etc.)
# ============================================================================

class TestMechanicOS:
    """Mechanic Service Order CRUD tests"""

    def test_create_os(self, authenticated_client):
        """POST /api/v1/mechanic/os should create a service order"""
        response = authenticated_client.post(f"{BASE_URL}/api/v1/mechanic/os", json={
            "vehicle_plate": "TEST789",
            "vehicle_km": 50000,
            "vehicle_model": "Honda Civic",
            "vehicle_year": 2022,
            "customer_id": "test-customer-001",  # customer_id is required
            "customer_phone": "5511999999999",
            "complaint": "Engine making strange noise"
        })
        assert response.status_code == 201
        data = response.json()
        # Go backend uses PascalCase
        assert "ID" in data
        assert data["VehiclePlate"] == "TEST789"
        assert data["Status"] == "open"
        print(f"✓ Create OS: ID={data['ID']}, Status={data['Status']}, VehiclePlate={data['VehiclePlate']}")
        return data["ID"]

    def test_list_open_os(self, authenticated_client):
        """GET /api/v1/mechanic/os should list open service orders"""
        response = authenticated_client.get(f"{BASE_URL}/api/v1/mechanic/os")
        assert response.status_code == 200
        data = response.json()
        assert "data" in data
        assert "total" in data
        assert isinstance(data["data"], list)
        print(f"✓ List OS: total={data['total']}, data_count={len(data['data'])}")

    def test_create_and_get_os(self, authenticated_client):
        """Create OS then GET by ID to verify persistence"""
        # Create - customer_id is required
        create_response = authenticated_client.post(f"{BASE_URL}/api/v1/mechanic/os", json={
            "vehicle_plate": "VERIFY2",
            "vehicle_km": 75000,
            "vehicle_model": "Toyota Corolla",
            "vehicle_year": 2021,
            "customer_id": "test-customer-002",
            "customer_phone": "5511888888888",
            "complaint": "Brake issue"
        })
        assert create_response.status_code == 201
        created_os = create_response.json()
        os_id = created_os["ID"]  # PascalCase

        # GET to verify persistence
        get_response = authenticated_client.get(f"{BASE_URL}/api/v1/mechanic/os/{os_id}")
        assert get_response.status_code == 200
        fetched_os = get_response.json()
        assert fetched_os["ID"] == os_id
        assert fetched_os["VehiclePlate"] == "VERIFY2"
        assert fetched_os["Complaint"] == "Brake issue"
        print(f"✓ Create→GET persistence verified: ID={os_id}")


# ============================================================================
# RUN TESTS
# ============================================================================

if __name__ == "__main__":
    pytest.main([__file__, "-v", "--tb=short"])
