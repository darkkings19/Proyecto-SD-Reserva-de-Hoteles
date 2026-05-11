from dataclasses import dataclass
from datetime import datetime
from typing import Optional

@dataclass
class Notification:
    user_id: str
    reservation_id: str
    tipo: str
    created_at: Optional[datetime] = None
