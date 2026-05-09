from dataclasses import dataclass

@dataclass
class Notification:
    user_id: str
    reservation_id: str
    tipo: str
