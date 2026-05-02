const BASE_URL = 'http://127.0.0.1:9876'

let printersList = []

async function checkStatus() {
  const dot = document.getElementById('status-dot')
  const text = document.getElementById('status-text')

  try {
    const res = await fetch(`${BASE_URL}/status`)
    const data = await res.json()
    dot.className = 'w-2 h-2 rounded-full bg-green-500'
    text.textContent = data.message
  } catch {
    dot.className = 'w-2 h-2 rounded-full bg-red-400'
    text.textContent = 'Sin conexión con el agente'
  }
}

async function fetchPrinters() {
  try {
    const res = await fetch(`${BASE_URL}/printers`)
    const data = await res.json()
    printersList = data.printers || []
    populatePrinterSelects()
  } catch {
    printersList = []
  }
}

function populatePrinterSelects() {
  const selects = ['pdf-printer', 'zpl-printer']

  selects.forEach(id => {
    const select = document.getElementById(id)
    select.innerHTML = ''

    if (printersList.length === 0) {
      select.innerHTML = '<option value="">Sin impresoras disponibles</option>'
      return
    }

    printersList.forEach(p => {
      const option = document.createElement('option')
      option.value = p.name
      option.textContent = p.name + (p.is_default ? ' (predeterminada)' : '')
      if (p.is_default) option.selected = true
      select.appendChild(option)
    })
  })
}

async function loadPrinters() {
  await fetchPrinters()

  const container = document.getElementById('printers-list')
  container.innerHTML = ''
  container.classList.remove('hidden')

  if (printersList.length === 0) {
    container.innerHTML = '<p class="text-sm text-slate-400">No se encontraron impresoras.</p>'
    return
  }

  const grid = document.createElement('div')
  grid.className = 'grid grid-cols-1 md:grid-cols-2 gap-2'

  printersList.forEach(p => {
    const el = document.createElement('div')
    el.className = 'flex items-center justify-between bg-slate-50 border border-slate-200 rounded-xl px-4 py-3'
    el.innerHTML = `
      <div>
        <p class="text-sm font-medium text-slate-800">${p.name}</p>
        <p class="text-xs text-slate-400">${p.driver} — ${p.port}</p>
      </div>
      ${p.is_default ? '<span class="text-xs bg-indigo-100 text-indigo-700 px-2 py-0.5 rounded-full font-medium">Predeterminada</span>' : ''}
    `
    grid.appendChild(el)
  })

  container.appendChild(grid)
}

async function printPDF() {
  const file = document.getElementById('pdf-file').files[0]
  const select = document.getElementById('pdf-printer')
  const useDefault = document.getElementById('pdf-default').checked
  const copies = parseInt(document.getElementById('pdf-copies').value) || 1
  const resultEl = document.getElementById('pdf-result')

  if (!file) {
    showResult(resultEl, false, 'Selecciona un archivo PDF.')
    return
  }

  const printerName = useDefault ? '' : select.value

  if (!useDefault && !printerName) {
    showResult(resultEl, false, 'Selecciona una impresora.')
    return
  }

  const base64 = await fileToBase64(file)

  await sendPrint('/print/pdf', {
    printer_name: printerName,
    use_default_printer: useDefault,
    pdf_base64: base64,
    copies: copies
  }, resultEl)
}

async function sendZPL() {
  const select = document.getElementById('zpl-printer')
  const useDefault = document.getElementById('zpl-default').checked
  const raw = document.getElementById('zpl-raw').value.trim()
  const copies = parseInt(document.getElementById('zpl-copies').value) || 1
  const resultEl = document.getElementById('zpl-result')

  if (!raw) {
    showResult(resultEl, false, 'El comando ZPL es requerido.')
    return
  }

  const printerName = useDefault ? '' : select.value

  if (!useDefault && !printerName) {
    showResult(resultEl, false, 'Selecciona una impresora.')
    return
  }

  await sendPrint('/print/zpl/raw', {
    printer_name: printerName,
    use_default_printer: useDefault,
    raw: raw,
    copies: copies
  }, resultEl)
}

async function sendPrint(endpoint, body, resultEl) {
  try {
    const res = await fetch(`${BASE_URL}${endpoint}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body)
    })

    const data = await res.json()

    if (res.ok) {
      showResult(resultEl, true, data.message)
    } else {
      showResult(resultEl, false, data.error)
    }
  } catch {
    showResult(resultEl, false, 'No se pudo conectar al agente.')
  }
}

function fileToBase64(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(reader.result.split(',')[1])
    reader.onerror = reject
    reader.readAsDataURL(file)
  })
}

function showResult(el, success, message) {
  el.className = `mt-1 px-4 py-3 rounded-lg text-sm font-medium ${success ? 'bg-green-50 text-green-700 border border-green-200' : 'bg-red-50 text-red-700 border border-red-200'}`
  el.textContent = message
}

function togglePrinterSelect(selectId, checkboxId) {
  const select = document.getElementById(selectId)
  const checked = document.getElementById(checkboxId).checked
  select.disabled = checked
  select.classList.toggle('opacity-40', checked)

  if (checked) {
    const defaultPrinter = printersList.find(p => p.is_default)
    if (defaultPrinter) {
      select.value = defaultPrinter.name
    }
  }
}

function updateFilename(inputId, labelId) {
  const input = document.getElementById(inputId)
  const label = document.getElementById(labelId)
  label.textContent = input.files[0] ? input.files[0].name : 'Seleccionar archivo...'
  label.classList.toggle('text-slate-800', !!input.files[0])
  label.classList.toggle('text-slate-400', !input.files[0])
}

checkStatus()
fetchPrinters()