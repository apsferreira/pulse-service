"""
pulse-iit — SDK Python server-side para pulse-service IIT
BKL-1004: event tracking com buffer thread-safe e batch flush

Uso minimo:
    from pulse_iit import PulseClient

    pulse = PulseClient(project_id="proj_xxx", api_url="https://pulse-api...")
    pulse.track("checkout_started", {"cart_value": 99.9}, user_id="user_123")
    pulse.shutdown()
"""

from .client import PulseClient

__all__ = ["PulseClient"]
__version__ = "0.1.0"
