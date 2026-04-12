"""
PulseClient — cliente Python server-side para o pulse-service IIT.

Buffer thread-safe com flush automatico a cada 5s ou ao atingir 50 eventos.
Flush roda em thread dedicada — nao bloqueia o thread principal da aplicacao.
"""

from __future__ import annotations

import json
import logging
import threading
import time
import urllib.request
import urllib.error
from datetime import datetime, timezone
from typing import Any, Dict, List, Optional

logger = logging.getLogger(__name__)

# Tipos de evento
_TYPE_TRACK = "track"
_TYPE_IDENTIFY = "identify"


def _now_iso() -> str:
    return datetime.now(tz=timezone.utc).isoformat()


class PulseClient:
    """
    Cliente Python server-side para o pulse-service IIT.

    Parametros
    ----------
    project_id : str
        ID do projeto no pulse-service.
    api_url : str
        URL base do pulse-service
        (ex: "https://pulse-api.institutoitinerante.com.br").
    flush_interval : float
        Intervalo de flush automatico em segundos (default: 5.0).
    max_batch_size : int
        Numero maximo de eventos por batch antes de flush forcado (default: 50).
    request_timeout : float
        Timeout por requisicao HTTP em segundos (default: 10.0).
    """

    def __init__(
        self,
        project_id: str,
        api_url: str,
        flush_interval: float = 5.0,
        max_batch_size: int = 50,
        request_timeout: float = 10.0,
    ) -> None:
        if not project_id:
            raise ValueError("project_id e obrigatorio")
        if not api_url:
            raise ValueError("api_url e obrigatorio")

        self.project_id = project_id
        self.api_url = api_url.rstrip("/")
        self.flush_interval = flush_interval
        self.max_batch_size = max_batch_size
        self.request_timeout = request_timeout

        self._buffer: List[Dict[str, Any]] = []
        self._lock = threading.Lock()
        self._shutdown_event = threading.Event()
        self._flushing = False

        # Thread de flush automatico — daemon para nao bloquear o processo
        self._flush_thread = threading.Thread(
            target=self._flush_loop,
            name="pulse-flush",
            daemon=True,
        )
        self._flush_thread.start()

    # --- Public API ---

    def track(
        self,
        event: str,
        properties: Optional[Dict[str, Any]] = None,
        user_id: Optional[str] = None,
    ) -> None:
        """
        Registra um evento de tracking no buffer.

        O buffer e enviado automaticamente a cada `flush_interval` segundos
        ou quando atinge `max_batch_size` eventos.

        Parametros
        ----------
        event : str
            Nome do evento (ex: "checkout_started", "page_view").
        properties : dict, optional
            Propriedades adicionais do evento.
        user_id : str, optional
            Identificador do usuario.
        """
        if self._shutdown_event.is_set():
            logger.warning("track() chamado apos shutdown() — evento descartado")
            return

        payload: Dict[str, Any] = {
            "type": _TYPE_TRACK,
            "event": event,
            "project_id": self.project_id,
            "timestamp": _now_iso(),
        }
        if user_id is not None:
            payload["user_id"] = user_id
        if properties:
            payload["properties"] = properties

        self._enqueue(payload)

    def identify(
        self,
        user_id: str,
        traits: Optional[Dict[str, Any]] = None,
    ) -> None:
        """
        Registra um evento de identificacao de usuario no buffer.

        Parametros
        ----------
        user_id : str
            Identificador do usuario.
        traits : dict, optional
            Atributos do usuario (ex: {"plan": "pro", "company": "IIT"}).
        """
        if self._shutdown_event.is_set():
            logger.warning("identify() chamado apos shutdown() — evento descartado")
            return

        payload: Dict[str, Any] = {
            "type": _TYPE_IDENTIFY,
            "user_id": user_id,
            "project_id": self.project_id,
            "timestamp": _now_iso(),
        }
        if traits:
            payload["traits"] = traits

        self._enqueue(payload)

    def flush(self) -> None:
        """
        Envia todos os eventos em buffer para o pulse-service imediatamente.

        Bloqueia ate a requisicao HTTP completar (ou timeout).
        Thread-safe — pode ser chamado de qualquer thread.
        """
        with self._lock:
            if not self._buffer:
                return
            batch = self._buffer[:]
            self._buffer.clear()

        self._send_batch(batch)

    def shutdown(self) -> None:
        """
        Flush final do buffer e encerra a thread de flush automatico.

        Deve ser chamado no encerramento da aplicacao (SIGTERM/SIGINT).
        Bloqueia ate o flush ser concluido.
        """
        self._shutdown_event.set()
        self._flush_thread.join(timeout=self.request_timeout + 2)
        self.flush()

    # --- Private ---

    def _enqueue(self, event: Dict[str, Any]) -> None:
        with self._lock:
            self._buffer.append(event)
            should_flush = len(self._buffer) >= self.max_batch_size

        if should_flush:
            # Flush em thread separada para nao bloquear o chamador
            threading.Thread(
                target=self.flush,
                name="pulse-flush-forced",
                daemon=True,
            ).start()

    def _flush_loop(self) -> None:
        """Loop executado na thread daemon — flush a cada `flush_interval` segundos."""
        while not self._shutdown_event.wait(timeout=self.flush_interval):
            try:
                self.flush()
            except Exception:
                logger.exception("pulse: erro inesperado no flush_loop")

    def _send_batch(self, events: List[Dict[str, Any]]) -> None:
        """Envia o batch para POST /api/v1/events/batch."""
        body = json.dumps(
            {"project_id": self.project_id, "events": events}
        ).encode("utf-8")

        req = urllib.request.Request(
            url=f"{self.api_url}/api/v1/events/batch",
            data=body,
            headers={"Content-Type": "application/json"},
            method="POST",
        )

        try:
            with urllib.request.urlopen(req, timeout=self.request_timeout) as resp:
                if resp.status not in (200, 201, 202):
                    logger.error(
                        "pulse: resposta inesperada %s ao enviar %d eventos",
                        resp.status,
                        len(events),
                    )
        except urllib.error.HTTPError as exc:
            logger.error(
                "pulse: HTTP %s ao enviar batch de %d eventos — %s",
                exc.code,
                len(events),
                exc.reason,
            )
        except urllib.error.URLError as exc:
            logger.error(
                "pulse: erro de conexao ao enviar batch de %d eventos — %s",
                len(events),
                exc.reason,
            )
        except Exception:
            logger.exception(
                "pulse: erro inesperado ao enviar batch de %d eventos", len(events)
            )
