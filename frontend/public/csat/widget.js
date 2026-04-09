(function () {
  'use strict';

  // BKL-344: Widget CSAT embeddable para coleta de NPS in-page.
  // Uso: <script src="...widget.js" data-survey-id="<uuid>" data-position="bottom-right"></script>

  var script = document.currentScript || (function () {
    var scripts = document.getElementsByTagName('script');
    return scripts[scripts.length - 1];
  })();

  var surveyId = script.getAttribute('data-survey-id');
  var position = script.getAttribute('data-position') || 'bottom-right';

  if (!surveyId) {
    console.warn('[pulse-csat] data-survey-id ausente — widget não iniciado.');
    return;
  }

  // Verificar se o usuário já respondeu (localStorage TTL 30 dias)
  var storageKey = 'pulse_csat_responded_' + surveyId;
  var storedTs = localStorage.getItem(storageKey);
  if (storedTs) {
    var diffDays = (Date.now() - parseInt(storedTs, 10)) / (1000 * 60 * 60 * 24);
    if (diffDays < 30) {
      return;
    }
  }

  // Determinar URL base a partir do atributo src do script
  var apiBase = (function () {
    var src = script.getAttribute('src') || '';
    var match = src.match(/^(https?:\/\/[^/]+)/);
    return match ? match[1] : '';
  })();

  // --- Estilos inline ---
  var COLORS = {
    primary: '#6366f1',
    primaryHover: '#4f46e5',
    bg: '#ffffff',
    overlay: 'rgba(0,0,0,0.4)',
    text: '#1f2937',
    muted: '#6b7280',
    border: '#e5e7eb',
    success: '#10b981',
  };

  var positionStyles = {
    'bottom-right': 'bottom:24px;right:24px;',
    'bottom-left': 'bottom:24px;left:24px;',
    'top-right': 'top:24px;right:24px;',
    'top-left': 'top:24px;left:24px;',
  };
  var posStyle = positionStyles[position] || positionStyles['bottom-right'];

  // --- Botão flutuante ---
  var btn = document.createElement('button');
  btn.id = 'pulse-csat-btn';
  btn.textContent = 'Avaliar';
  btn.setAttribute('aria-label', 'Abrir pesquisa de satisfação');
  btn.style.cssText = [
    'position:fixed;' + posStyle,
    'z-index:9998;',
    'background:' + COLORS.primary + ';',
    'color:#fff;',
    'border:none;',
    'border-radius:24px;',
    'padding:10px 20px;',
    'font-size:14px;',
    'font-family:system-ui,sans-serif;',
    'font-weight:600;',
    'cursor:pointer;',
    'box-shadow:0 4px 12px rgba(99,102,241,0.4);',
    'transition:background 0.2s;',
  ].join('');
  btn.onmouseover = function () { btn.style.background = COLORS.primaryHover; };
  btn.onmouseout = function () { btn.style.background = COLORS.primary; };
  document.body.appendChild(btn);

  // --- Overlay + Modal ---
  var overlay = document.createElement('div');
  overlay.id = 'pulse-csat-overlay';
  overlay.style.cssText = [
    'display:none;',
    'position:fixed;top:0;left:0;width:100%;height:100%;',
    'background:' + COLORS.overlay + ';',
    'z-index:9999;',
    'align-items:center;justify-content:center;',
  ].join('');

  var modal = document.createElement('div');
  modal.style.cssText = [
    'background:' + COLORS.bg + ';',
    'border-radius:12px;',
    'padding:28px;',
    'max-width:400px;width:90%;',
    'box-shadow:0 20px 60px rgba(0,0,0,0.2);',
    'font-family:system-ui,sans-serif;',
    'color:' + COLORS.text + ';',
  ].join('');

  // Pergunta
  var questionEl = document.createElement('p');
  questionEl.style.cssText = 'margin:0 0 16px;font-size:16px;font-weight:600;line-height:1.4;';
  questionEl.textContent = 'Em uma escala de 0 a 10, o quanto você recomendaria nosso serviço?';

  // Score buttons (0-10)
  var scoresContainer = document.createElement('div');
  scoresContainer.style.cssText = 'display:flex;gap:6px;flex-wrap:wrap;margin-bottom:20px;';

  var selectedScore = null;

  for (var i = 0; i <= 10; i++) {
    (function (score) {
      var scoreBtn = document.createElement('button');
      scoreBtn.textContent = String(score);
      scoreBtn.setAttribute('data-score', String(score));
      scoreBtn.style.cssText = [
        'width:36px;height:36px;',
        'border:2px solid ' + COLORS.border + ';',
        'border-radius:8px;',
        'background:#fff;',
        'color:' + COLORS.text + ';',
        'font-size:13px;font-weight:600;',
        'cursor:pointer;',
        'transition:all 0.15s;',
      ].join('');
      scoreBtn.onmouseover = function () {
        if (selectedScore !== score) {
          scoreBtn.style.borderColor = COLORS.primary;
          scoreBtn.style.color = COLORS.primary;
        }
      };
      scoreBtn.onmouseout = function () {
        if (selectedScore !== score) {
          scoreBtn.style.borderColor = COLORS.border;
          scoreBtn.style.color = COLORS.text;
        }
      };
      scoreBtn.onclick = function () {
        // Resetar todos
        var allBtns = scoresContainer.querySelectorAll('button');
        for (var j = 0; j < allBtns.length; j++) {
          allBtns[j].style.background = '#fff';
          allBtns[j].style.color = COLORS.text;
          allBtns[j].style.borderColor = COLORS.border;
        }
        selectedScore = score;
        scoreBtn.style.background = COLORS.primary;
        scoreBtn.style.color = '#fff';
        scoreBtn.style.borderColor = COLORS.primary;
      };
      scoresContainer.appendChild(scoreBtn);
    })(i);
  }

  // Labels de âncora
  var anchors = document.createElement('div');
  anchors.style.cssText = 'display:flex;justify-content:space-between;margin-bottom:16px;';
  var anchorLeft = document.createElement('span');
  anchorLeft.style.cssText = 'font-size:11px;color:' + COLORS.muted + ';';
  anchorLeft.textContent = 'Nada provável';
  var anchorRight = document.createElement('span');
  anchorRight.style.cssText = 'font-size:11px;color:' + COLORS.muted + ';';
  anchorRight.textContent = 'Muito provável';
  anchors.appendChild(anchorLeft);
  anchors.appendChild(anchorRight);

  // Comentário
  var commentLabel = document.createElement('label');
  commentLabel.style.cssText = 'display:block;font-size:13px;color:' + COLORS.muted + ';margin-bottom:6px;';
  commentLabel.textContent = 'Comentário (opcional)';
  var commentInput = document.createElement('textarea');
  commentInput.rows = 3;
  commentInput.placeholder = 'O que motivou sua nota?';
  commentInput.style.cssText = [
    'width:100%;box-sizing:border-box;',
    'border:1px solid ' + COLORS.border + ';',
    'border-radius:8px;padding:8px 10px;',
    'font-size:13px;font-family:system-ui,sans-serif;',
    'resize:vertical;margin-bottom:16px;',
    'outline:none;',
  ].join('');

  // Botão enviar
  var submitBtn = document.createElement('button');
  submitBtn.textContent = 'Enviar';
  submitBtn.style.cssText = [
    'width:100%;padding:10px;',
    'background:' + COLORS.primary + ';color:#fff;',
    'border:none;border-radius:8px;',
    'font-size:14px;font-weight:600;',
    'cursor:pointer;transition:background 0.2s;',
  ].join('');
  submitBtn.onmouseover = function () { submitBtn.style.background = COLORS.primaryHover; };
  submitBtn.onmouseout = function () { submitBtn.style.background = COLORS.primary; };

  // Mensagem de sucesso (oculta inicialmente)
  var successMsg = document.createElement('div');
  successMsg.style.cssText = [
    'display:none;text-align:center;padding:24px 0;',
    'font-size:16px;font-weight:600;color:' + COLORS.success + ';',
  ].join('');
  successMsg.textContent = 'Obrigado pelo feedback!';

  // Botão fechar (X)
  var closeBtn = document.createElement('button');
  closeBtn.textContent = '×';
  closeBtn.setAttribute('aria-label', 'Fechar pesquisa');
  closeBtn.style.cssText = [
    'position:absolute;top:12px;right:16px;',
    'background:none;border:none;font-size:22px;',
    'color:' + COLORS.muted + ';cursor:pointer;line-height:1;',
  ].join('');
  modal.style.position = 'relative';

  // Montar modal
  modal.appendChild(closeBtn);
  modal.appendChild(questionEl);
  modal.appendChild(scoresContainer);
  modal.appendChild(anchors);
  modal.appendChild(commentLabel);
  modal.appendChild(commentInput);
  modal.appendChild(submitBtn);
  modal.appendChild(successMsg);
  overlay.appendChild(modal);
  document.body.appendChild(overlay);

  // --- Lógica de abertura/fechamento ---
  function openModal() {
    overlay.style.display = 'flex';
    selectedScore = null;
    commentInput.value = '';
    submitBtn.style.display = 'block';
    successMsg.style.display = 'none';
    var allBtns = scoresContainer.querySelectorAll('button');
    for (var j = 0; j < allBtns.length; j++) {
      allBtns[j].style.background = '#fff';
      allBtns[j].style.color = COLORS.text;
      allBtns[j].style.borderColor = COLORS.border;
    }
  }

  function closeModal() {
    overlay.style.display = 'none';
  }

  btn.onclick = openModal;
  closeBtn.onclick = closeModal;
  overlay.onclick = function (e) {
    if (e.target === overlay) closeModal();
  };

  // --- Envio ---
  submitBtn.onclick = function () {
    if (selectedScore === null) {
      alert('Por favor, selecione uma nota antes de enviar.');
      return;
    }

    submitBtn.disabled = true;
    submitBtn.textContent = 'Enviando...';

    var payload = JSON.stringify({
      score: selectedScore,
      comment: commentInput.value.trim() || null,
    });

    var endpoint = apiBase + '/public/csat/' + surveyId + '/respond';

    var xhr = new XMLHttpRequest();
    xhr.open('POST', endpoint, true);
    xhr.setRequestHeader('Content-Type', 'application/json');
    xhr.onload = function () {
      submitBtn.style.display = 'none';
      successMsg.style.display = 'block';
      // Salvar timestamp no localStorage para suprimir por 30 dias
      localStorage.setItem(storageKey, String(Date.now()));
      // Fechar modal após 2s
      setTimeout(closeModal, 2000);
    };
    xhr.onerror = function () {
      // Mesmo em erro, marcar como respondido para não perturbar o usuário
      submitBtn.style.display = 'none';
      successMsg.style.display = 'block';
      localStorage.setItem(storageKey, String(Date.now()));
      setTimeout(closeModal, 2000);
    };
    xhr.send(payload);
  };
})();
