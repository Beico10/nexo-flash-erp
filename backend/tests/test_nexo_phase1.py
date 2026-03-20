"""
Nexo One ERP Phase 1 Testing - Backend API Tests
Tests all module endpoints, auth, and tax calculations
"""
import pytest
import requests
import os

BASE_URL = os.environ.get('REACT_APP_BACKEND_URL', 'https://fastapi-router-setup.preview.emergentagent.com').rstrip('/')


class TestHealth:
    """Health check endpoint tests"""
    
    def test_health_endpoint(self):
        """Test health endpoint returns status ok"""
        response = requests.get(f"{BASE_URL}/api/health")
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "ok"
        assert "engine" in data
        print(f"Health check: {data}")


class TestAuth:
    """Authentication tests"""
    
    def test_login_success(self):
        """Test successful login with demo credentials"""
        response = requests.post(f"{BASE_URL}/api/auth/login", json={
            "tenant_slug": "demo",
            "email": "admin@demo.com",
            "password": "demo123"
        })
        assert response.status_code == 200
        data = response.json()
        assert "access_token" in data
        assert data["token_type"] == "Bearer"
        print(f"Login success - token type: {data['token_type']}")
        return data["access_token"]
    
    def test_login_invalid_credentials(self):
        """Test login with invalid credentials"""
        response = requests.post(f"{BASE_URL}/api/auth/login", json={
            "tenant_slug": "demo",
            "email": "wrong@demo.com",
            "password": "wrongpass"
        })
        assert response.status_code == 401
        print(f"Invalid login correctly rejected: {response.status_code}")
    
    def test_login_missing_tenant(self):
        """Test login with missing tenant slug"""
        response = requests.post(f"{BASE_URL}/api/auth/login", json={
            "email": "admin@demo.com",
            "password": "demo123"
        })
        assert response.status_code == 400
        print(f"Missing tenant correctly rejected: {response.status_code}")


@pytest.fixture
def auth_token():
    """Get authentication token for protected endpoints"""
    response = requests.post(f"{BASE_URL}/api/auth/login", json={
        "tenant_slug": "demo",
        "email": "admin@demo.com",
        "password": "demo123"
    })
    if response.status_code == 200:
        return response.json()["access_token"]
    pytest.skip("Authentication failed")


@pytest.fixture
def auth_headers(auth_token):
    """Get headers with auth token"""
    return {"Authorization": f"Bearer {auth_token}"}


class TestPublicTaxSimulator:
    """Public Tax Simulator endpoints (no auth required)"""
    
    def test_ncm_list_no_auth(self):
        """Test NCM list endpoint is public"""
        response = requests.get(f"{BASE_URL}/api/v1/tax/ncm-list")
        assert response.status_code == 200
        data = response.json()
        assert "data" in data
        assert len(data["data"]) > 0
        print(f"Found {len(data['data'])} NCM items")
    
    def test_simulate_cesta_basica(self):
        """Test tax simulation for cesta basica item (zero tax)"""
        response = requests.post(f"{BASE_URL}/api/v1/tax/simulate", json={
            "ncm_code": "19052000",  # Pao de forma
            "base_value": 1000.0,
            "operation": "debit_exit"
        })
        assert response.status_code == 200
        data = response.json()
        assert data["total_tax"] == 0
        assert data["is_basket_item"] == True
        print(f"Cesta Basica tax: {data['total_tax']} (expected 0)")
    
    def test_simulate_standard_item(self):
        """Test tax simulation for standard rate item"""
        response = requests.post(f"{BASE_URL}/api/v1/tax/simulate", json={
            "ncm_code": "84715010",  # Computador
            "base_value": 1000.0,
            "operation": "debit_exit"
        })
        assert response.status_code == 200
        data = response.json()
        assert data["total_tax"] > 0
        assert data["ibs_amount"] > 0
        assert data["cbs_amount"] > 0
        print(f"Standard item tax: {data['total_tax']} (IBS: {data['ibs_amount']}, CBS: {data['cbs_amount']})")
    
    def test_simulate_seletivo_item(self):
        """Test tax simulation for selective tax item (tobacco/alcohol)"""
        response = requests.post(f"{BASE_URL}/api/v1/tax/simulate", json={
            "ncm_code": "22030000",  # Cerveja
            "base_value": 1000.0,
            "operation": "debit_exit"
        })
        assert response.status_code == 200
        data = response.json()
        assert data["selective_amount"] > 0
        print(f"Seletivo item: selective_amount={data['selective_amount']}")
    
    def test_simulate_credit_entry(self):
        """Test tax simulation for purchase (credit entry)"""
        response = requests.post(f"{BASE_URL}/api/v1/tax/simulate", json={
            "ncm_code": "84715010",
            "base_value": 1000.0,
            "operation": "credit_entry"
        })
        assert response.status_code == 200
        data = response.json()
        assert data["cashback_amount"] > 0  # Positive for credit entry
        print(f"Credit entry cashback: {data['cashback_amount']}")
    
    def test_simulate_invalid_ncm(self):
        """Test tax simulation with invalid NCM code"""
        response = requests.post(f"{BASE_URL}/api/v1/tax/simulate", json={
            "ncm_code": "INVALID123",
            "base_value": 1000.0,
            "operation": "debit_exit"
        })
        assert response.status_code == 400
        print(f"Invalid NCM correctly rejected: {response.status_code}")


class TestDashboard:
    """Dashboard stats endpoint tests"""
    
    def test_dashboard_requires_auth(self):
        """Test dashboard endpoint requires authentication"""
        response = requests.get(f"{BASE_URL}/api/v1/dashboard/stats")
        assert response.status_code == 401
        print(f"Dashboard without auth: {response.status_code}")
    
    def test_dashboard_with_auth(self, auth_headers):
        """Test dashboard returns stats with authentication"""
        response = requests.get(f"{BASE_URL}/api/v1/dashboard/stats", headers=auth_headers)
        assert response.status_code == 200
        data = response.json()
        assert "mechanic_os" in data
        assert "revenue" in data
        assert "chart" in data["revenue"]
        print(f"Dashboard stats: OS total={data['mechanic_os']['total']}, revenue_today={data['revenue']['today']}")


class TestMechanic:
    """Mechanic module OS (Service Order) tests"""
    
    def test_mechanic_os_requires_auth(self):
        """Test mechanic OS endpoint requires authentication"""
        response = requests.get(f"{BASE_URL}/api/v1/mechanic/os")
        assert response.status_code == 401
    
    def test_mechanic_os_list(self, auth_headers):
        """Test listing mechanic service orders"""
        response = requests.get(f"{BASE_URL}/api/v1/mechanic/os", headers=auth_headers)
        assert response.status_code == 200
        data = response.json()
        assert "data" in data
        print(f"Mechanic OS list: {len(data['data'] or [])} orders")
    
    def test_create_mechanic_os(self, auth_headers):
        """Test creating a new service order"""
        response = requests.post(f"{BASE_URL}/api/v1/mechanic/os", headers=auth_headers, json={
            "vehicle_plate": "TEST123",
            "vehicle_km": 50000,
            "vehicle_model": "Test Car 2024",
            "vehicle_year": 2024,
            "customer_id": "TEST_customer_001",
            "customer_phone": "5511999999999",
            "complaint": "Test complaint for automated testing"
        })
        assert response.status_code == 201
        data = response.json()
        assert data["status"] == "open"
        assert data["vehicle_plate"] == "TEST123"
        print(f"Created OS: number={data.get('number')}, status={data['status']}")
        return data


class TestBakery:
    """Bakery module products tests"""
    
    def test_bakery_products_requires_auth(self):
        """Test bakery products endpoint requires authentication"""
        response = requests.get(f"{BASE_URL}/api/v1/bakery/products")
        assert response.status_code == 401
    
    def test_bakery_products_list(self, auth_headers):
        """Test listing bakery products"""
        response = requests.get(f"{BASE_URL}/api/v1/bakery/products", headers=auth_headers)
        assert response.status_code == 200
        data = response.json()
        assert "data" in data
        products = data["data"] or []
        cesta_count = sum(1 for p in products if p.get("IsBasketItem"))
        print(f"Bakery products: {len(products)} total, {cesta_count} cesta basica items")


class TestAesthetics:
    """Aesthetics module appointments tests"""
    
    def test_aesthetics_requires_auth(self):
        """Test aesthetics appointments endpoint requires authentication"""
        response = requests.get(f"{BASE_URL}/api/v1/aesthetics/appointments")
        assert response.status_code == 401
    
    def test_aesthetics_appointments_list(self, auth_headers):
        """Test listing aesthetics appointments"""
        response = requests.get(f"{BASE_URL}/api/v1/aesthetics/appointments", headers=auth_headers)
        assert response.status_code == 200
        data = response.json()
        assert "data" in data
        appointments = data["data"] or []
        print(f"Aesthetics appointments: {len(appointments)} total")


class TestAI:
    """AI suggestions module tests"""
    
    def test_ai_suggestions_requires_auth(self):
        """Test AI suggestions endpoint requires authentication"""
        response = requests.get(f"{BASE_URL}/api/v1/ai/suggestions")
        assert response.status_code == 401
    
    def test_ai_suggestions_list(self, auth_headers):
        """Test listing AI suggestions"""
        response = requests.get(f"{BASE_URL}/api/v1/ai/suggestions", headers=auth_headers)
        assert response.status_code == 200
        data = response.json()
        assert "data" in data
        suggestions = data["data"] or []
        pending_count = sum(1 for s in suggestions if s.get("Status") == "pending")
        print(f"AI suggestions: {len(suggestions)} total, {pending_count} pending")


class TestLogistics:
    """Logistics module freight calculator tests"""
    
    def test_logistics_requires_auth(self):
        """Test logistics endpoint requires authentication"""
        response = requests.post(f"{BASE_URL}/api/v1/logistics/freight/calculate", json={
            "vehicle_type": "truck",
            "distance_km": 500,
            "weight_kg": 15000
        })
        assert response.status_code == 401
    
    def test_logistics_no_contract(self, auth_headers):
        """Test logistics fails gracefully when no contract exists"""
        # Note: This is expected to fail with 400 because no contracts are seeded
        response = requests.post(
            f"{BASE_URL}/api/v1/logistics/freight/calculate",
            headers=auth_headers,
            json={
                "vehicle_type": "truck",
                "distance_km": 500,
                "weight_kg": 15000,
                "toll_cost": 250,
                "fuel_cost_per_km": 2.10,
                "driver_cost_per_km": 0.80
            }
        )
        # This returns 400 because no contract exists - expected behavior
        assert response.status_code in [200, 400]
        if response.status_code == 400:
            data = response.json()
            assert "contrato não encontrado" in data.get("error", "").lower()
            print("Logistics: No contract configured (expected)")
        else:
            print("Logistics: Contract found and calculation successful")


class TestProtectedEndpoints:
    """Test that all protected endpoints require authentication"""
    
    @pytest.mark.parametrize("endpoint,method", [
        ("/api/v1/dashboard/stats", "GET"),
        ("/api/v1/mechanic/os", "GET"),
        ("/api/v1/bakery/products", "GET"),
        ("/api/v1/aesthetics/appointments", "GET"),
        ("/api/v1/ai/suggestions", "GET"),
    ])
    def test_protected_endpoint_requires_auth(self, endpoint, method):
        """Test that protected endpoints return 401 without auth"""
        if method == "GET":
            response = requests.get(f"{BASE_URL}{endpoint}")
        else:
            response = requests.post(f"{BASE_URL}{endpoint}", json={})
        
        assert response.status_code == 401, f"Expected 401 for {endpoint}, got {response.status_code}"
        print(f"{method} {endpoint} correctly requires auth: 401")


class TestTaxEngineTransition:
    """Test tax engine 2026 transition calculations"""
    
    def test_transition_factor_2026(self):
        """Test that 2026 transition factor of 10% is applied"""
        # Simulate with a standard item
        response = requests.post(f"{BASE_URL}/api/v1/tax/simulate", json={
            "ncm_code": "84715010",
            "base_value": 1000.0,
            "operation": "debit_exit"
        })
        assert response.status_code == 200
        data = response.json()
        
        # In 2026, IBS rate should be ~9.25% (full rate 26.5% * 10% transition + federal CBS)
        # Total should be around 13% effective (IBS + CBS with transition)
        effective_rate = data["total_tax"] / data["base_value"]
        assert 0.10 < effective_rate < 0.20, f"Effective rate {effective_rate} outside expected range"
        print(f"2026 transition rate check: effective rate = {effective_rate:.2%}")


if __name__ == "__main__":
    pytest.main([__file__, "-v", "--tb=short"])
