import React, { useEffect, useMemo, useState } from 'react'

const defaults = {
  vcpu: 1, memory_gb: 1, concurrency: 80,
  avg_duration_ms: 200, requests_per_min: 600,
  region: 'asia-south1', min_instances: 0, max_instances: 5,
  idle_utilization_pc: 10
}

const LS_KEY = 'cihub_history_v1' // bump version if structure changes
const clamp = (n,min,max)=>Math.max(min,Math.min(max,n))

function InfinityLoop({size=260}) {
  const stroke = 'var(--accent)'
  const glow = '0 0 24px rgba(155,92,255,.8)'
  const d = `M ${size*0.15} ${size*0.5}
             C ${size*0.15} ${size*0.3}, ${size*0.35} ${size*0.3}, ${size*0.5} ${size*0.5}
             C ${size*0.65} ${size*0.7}, ${size*0.85} ${size*0.7}, ${size*0.85} ${size*0.5}
             C ${size*0.85} ${size*0.3}, ${size*0.65} ${size*0.3}, ${size*0.5} ${size*0.5}
             C ${size*0.35} ${size*0.7}, ${size*0.15} ${size*0.7}, ${size*0.15} ${size*0.5} Z`
  return (
    <svg width={size} height={size} viewBox={`0 0 ${size} ${size}`}>
      <path d={d} fill="none" stroke={stroke} strokeWidth="14" style={{filter:`drop-shadow(${glow})`}}/>
    </svg>
  )
}

const fmtMoney = v => typeof v==='number' ? `$${v.toFixed(4)}` : '-'
const fmtTime = ts => new Date(ts).toLocaleTimeString([], {hour:'2-digit', minute:'2-digit'})
const badge = (txt, tone='var(--accent)') => (
  <span style={{padding:'4px 10px', borderRadius:999, border:`1px solid ${tone}`, color:tone, fontSize:12}}>
    {txt}
  </span>
)

const Card = ({title, right, children}) => (
  <div style={{background:'var(--panel)', border:'1px solid var(--line)', borderRadius:16, padding:16}}>
    <div style={{display:'flex', justifyContent:'space-between', alignItems:'center', marginBottom:10}}>
      <strong style={{color:'var(--ink)'}}>{title}</strong>
      {right}
    </div>
    <div>{children}</div>
  </div>
)

export default function App() {
  const [form, setForm] = useState(defaults)
  const [resp, setResp] = useState(null)
  const [loading, setLoading] = useState(false)
  const [history, setHistory] = useState([])

  // Load history once
  useEffect(() => {
    try {
      const raw = localStorage.getItem(LS_KEY)
      if (raw) setHistory(JSON.parse(raw))
    } catch {}
  }, [])

  // Persist history on change
  useEffect(() => {
    try { localStorage.setItem(LS_KEY, JSON.stringify(history)) } catch {}
  }, [history])

  const riskTone = useMemo(() => {
    const s = resp?.risk_score ?? 0
    if (s >= 90) return 'var(--good)'
    if (s >= 70) return 'var(--warn)'
    return 'var(--bad)'
  }, [resp])

  const onChange = e => setForm({ ...form, [e.target.name]: e.target.value })

  const estimate = async () => {
    setLoading(true); setResp(null)
    try {
      const body = {
        ...form,
        vcpu: parseFloat(form.vcpu),
        memory_gb: parseFloat(form.memory_gb),
        concurrency: parseInt(form.concurrency),
        avg_duration_ms: parseInt(form.avg_duration_ms),
        requests_per_min: parseInt(form.requests_per_min),
        min_instances: parseInt(form.min_instances),
        max_instances: parseInt(form.max_instances),
        idle_utilization_pc: parseFloat(form.idle_utilization_pc)
      }
      const r = await fetch('/api/estimate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body)
      })
      const j = await r.json()
      setResp(j)

      // Save to local history (top, max 10)
      const entry = {
        ts: Date.now(),
        region: body.region,
        input: body,
        data: j,
        score: j?.risk_score ?? 0,
        costPer1k: j?.per_1k_requests?.cost_usd ?? 0
      }
      setHistory(h => [entry, ...h].slice(0, 10))
    } catch (e) {
      setResp({ error: String(e) })
    } finally { setLoading(false) }
  }

  const restore = (entry) => {
    // Restore result and inputs from a history pill
    if (entry?.input) setForm(entry.input)
    if (entry?.data) setResp(entry.data)
  }

  const Field = ({ label, name, type='number', step='any' }) => (
    <label style={{display:'block'}}>
      <div style={{fontSize:12, opacity:.7, marginBottom:6}}>{label}</div>
      <input
        name={name}
        value={form[name]}
        onChange={onChange}
        type={type}
        step={step}
        style={{
          padding:10, width:'100%',
          background:'var(--soft)', color:'var(--ink)',
          border:'1px solid var(--muter)', borderRadius:10, outline:'none'
        }}
      />
    </label>
  )

  return (
    <div style={{minHeight:'100vh', background:'var(--bg)', color:'var(--muted)'}}>
      <header style={{display:'flex', alignItems:'center', justifyContent:'space-between',
                      padding:'18px 22px', borderBottom:'1px solid var(--line)', position:'sticky', top:0, background:'var(--bg)'}}>
        <div style={{display:'flex', alignItems:'center', gap:12}}>
          {badge('DevOps ∞')}
          <span style={{fontWeight:700, color:'var(--ink)'}}>DevOps Intelligence Hub</span>
        </div>
        <div style={{opacity:.85}}>Automated Cloud Optimization for CI/CD</div>
      </header>

      <main style={{maxWidth:1200, margin:'0 auto', padding:'28px 22px'}}>
        {/* Hero with Infinity */}
        <div style={{display:'grid', gridTemplateColumns:'360px 1fr', gap:28, alignItems:'center', marginBottom:28}}>
          <div style={{display:'flex', justifyContent:'center'}}><InfinityLoop size={260} /></div>
          <div>
            <h2 style={{margin:'6px 0 8px', color:'var(--ink)'}}>Optimize every deploy</h2>
            <p style={{margin:0, opacity:.85}}>
              Estimate Cloud Run cost & CO₂, generate optimized YAML, forecast monthly bill,
              and enforce a single <b>Risk Score</b> as your CI/CD quality gate.
            </p>
            <div style={{marginTop:12, display:'flex', gap:10}}>
              {badge('CI → CD')}
              {badge('FinOps')}
              {badge('Sustainability')}
            </div>
          </div>
        </div>

        {/* Form + Results */}
        <div style={{display:'grid', gridTemplateColumns:'420px 1fr', gap:20}}>
          <Card
            title="Deployment Inputs"
            right={<span style={{opacity:.7, fontSize:12}}>WSL / EC2 friendly</span>}
          >
            <div style={{display:'grid', gridTemplateColumns:'repeat(2,1fr)', gap:12}}>
              <Field label="vCPU" name="vcpu"/>
              <Field label="Memory (GiB)" name="memory_gb"/>
              <Field label="Concurrency" name="concurrency"/>
              <Field label="Avg Duration (ms)" name="avg_duration_ms"/>
              <Field label="Requests / Min" name="requests_per_min"/>
              <Field label="Region" name="region" type="text"/>
              <Field label="Min Instances" name="min_instances"/>
              <Field label="Max Instances" name="max_instances"/>
              <Field label="Idle CPU % (min instances)" name="idle_utilization_pc"/>
            </div>
            <button onClick={estimate} disabled={loading}
              style={{marginTop:16, padding:'12px 16px', borderRadius:12, border:'1px solid var(--accent)',
                      background:'#100c18', color:'var(--ink)', cursor:'pointer', boxShadow:'0 0 16px rgba(155,92,255,.35)'}}>
              {loading ? 'Estimating…' : 'Estimate • Run Quality Gate'}
            </button>
          </Card>

          <div style={{display:'grid', gap:20}}>
            <Card title="Quality Gate">
              {!resp && <div style={{opacity:.7}}>Run an estimate to see Risk Score & advice.</div>}
              {resp && resp.error && <pre style={{color:'crimson'}}>{resp.error}</pre>}
              {resp && !resp.error && (
                <div style={{display:'grid', gridTemplateColumns:'repeat(3,1fr)', gap:12}}>
                  <div style={{background:'#0a0d12', border:'1px solid var(--line)', borderRadius:12, padding:12}}>
                    <div style={{fontSize:12, opacity:.7}}>Risk Score</div>
                    <div style={{fontSize:32, fontWeight:800, color:'var(--ink)'}}>
                      <span style={{color:riskTone}}>{resp.risk_score}</span>
                    </div>
                    <div style={{fontSize:12, opacity:.7}}>≥ 90 Production Ready</div>
                  </div>
                  <div style={{background:'#0a0d12', border:'1px solid var(--line)', borderRadius:12, padding:12}}>
                    <div style={{fontSize:12, opacity:.7}}>Cost / 1k req</div>
                    <div style={{fontSize:28}}>{fmtMoney(resp.per_1k_requests?.cost_usd)}</div>
                    <div style={{fontSize:12, opacity:.7}}>CO₂ / 1k: {Math.round(resp.per_1k_requests?.co2_g ?? 0)} g</div>
                  </div>
                  <div style={{background:'#0a0d12', border:'1px solid var(--line)', borderRadius:12, padding:12}}>
                    <div style={{fontSize:12, opacity:.7}}>Monthly Forecast</div>
                    <div style={{fontSize:28}}>${(resp.monthly_forecast?.cost_usd ?? 0).toFixed?.(2)}</div>
                    <div style={{fontSize:12, opacity:.7}}>{resp.monthly_forecast?.co2_kg ?? 0} kg CO₂</div>
                  </div>
                </div>
              )}
            </Card>

            <Card title="Suggested Cloud Run YAML">
              {!resp && <div style={{opacity:.7}}>Run estimate to generate YAML.</div>}
              {resp && !resp.error && (
                <pre style={{whiteSpace:'pre-wrap', background:'#0a0d12', padding:12, borderRadius:12, border:'1px solid var(--line)'}}>
{resp.suggested_yaml}
                </pre>
              )}
            </Card>

            <Card
              title="Advisor Notes"
              right={resp && !resp.error && <span style={{fontSize:12, opacity:.7}}>loaded • {new Date().toLocaleTimeString()}</span>}
            >
              {!resp && <div style={{opacity:.7}}>Suggestions will appear here.</div>}
              {resp && !resp.error && (
                <ul style={{margin:0, paddingLeft:18}}>
                  {resp.advice?.map((a,i)=><li key={i} style={{margin:'6px 0'}}>{a}</li>)}
                </ul>
              )}
            </Card>
          </div>
        </div>

        {/* History: lightning pills */}
        <div style={{marginTop:26}}>
          <Card title="Recent Runs (click to recall)">
            {!history?.length && <div style={{opacity:.7}}>No history yet — run an estimate.</div>}
            {!!history?.length && (
              <div style={{display:'flex', flexWrap:'wrap', gap:10}}>
                {history.map((h,idx) => (
                  <button key={h.ts}
                    onClick={()=>restore(h)}
                    title={`Restore ${fmtTime(h.ts)} • ${h.region}`}
                    style={{
                      cursor:'pointer',
                      borderRadius:999, border:'1px solid var(--line)',
                      background:'#0a0d12', color:'var(--ink)',
                      padding:'8px 12px', display:'flex', alignItems:'center', gap:8
                    }}>
                    <strong style={{color: (h.score>=90?'var(--good)': (h.score>=70?'var(--warn)':'var(--bad)'))}}>
                      {clamp(h.score,0,100)}
                    </strong>
                    <span>⚡</span>
                    <span style={{opacity:.9}}>{h.region}</span>
                    <span style={{opacity:.6}}>•</span>
                    <span style={{opacity:.85}}>{fmtMoney(h.costPer1k)}</span>
                    <span style={{opacity:.45, fontSize:12}}>• {fmtTime(h.ts)}</span>
                  </button>
                ))}
              </div>
            )}
          </Card>
        </div>
      </main>

      <footer style={{padding:'20px 22px', borderTop:'1px solid var(--line)', opacity:.6}}>
        © DevOps Intelligence Hub — CI • CD • Optimize • Monitor • Improve
      </footer>
    </div>
  )
}

