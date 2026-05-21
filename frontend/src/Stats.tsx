import * as React from 'react'

type WeightLog = {
    ID?: number
    id?: number
    Weight?: number
    weight?: number
    Date?: string
    date?: string
    CreatedAt?: string
    createdAt?: string
}

type Progress = {
    currentWeight: number
    startWeight: number
    targetWeight: number
    change: number
    logsCount: number
}

const defaultTelegramId = 898030333

export default function Stats() {
    const tgUser = (window as any).Telegram?.WebApp?.initDataUnsafe?.user
    const telegramId = tgUser?.id ?? defaultTelegramId

    const [weight, setWeight] = React.useState('')
    const [logs, setLogs] = React.useState<WeightLog[]>([])
    const [progress, setProgress] = React.useState<Progress | null>(null)
    const [loading, setLoading] = React.useState(true)
    const [saving, setSaving] = React.useState(false)

    async function loadStats() {
        setLoading(true)

        try {
            const [logsRes, progressRes] = await Promise.all([
                fetch(`/api/weight/${telegramId}`),
                fetch(`/api/progress/${telegramId}`),
            ])

            const logsData = await logsRes.json()
            const progressData = await progressRes.json()

            setLogs(Array.isArray(logsData) ? logsData : [])
            setProgress(progressData)
        } finally {
            setLoading(false)
        }
    }

    React.useEffect(() => {
        loadStats()
    }, [telegramId])

    async function saveWeight() {
        const value = Number(weight)

        if (!value || value < 20 || value > 300) {
            alert('Введите корректный вес от 20 до 300 кг')
            return
        }

        setSaving(true)

        try {
            const res = await fetch('/api/weight', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    telegramId,
                    weight: value,
                }),
            })

            if (!res.ok) {
                throw new Error(`Ошибка сохранения: ${res.status}`)
            }

            setWeight('')
            await loadStats()
        } catch (e) {
            alert(e instanceof Error ? e.message : 'Ошибка сохранения веса')
        } finally {
            setSaving(false)
        }
    }

    if (loading) {
        return (
            <div className="stack">
                <section className="card">
                    <h2>Статистика</h2>
                    <p className="muted">Загрузка прогресса...</p>
                </section>
            </div>
        )
    }

    const currentWeight = progress?.currentWeight ?? 0
    const startWeight = progress?.startWeight ?? 0
    const targetWeight = progress?.targetWeight ?? 0
    const change = progress?.change ?? 0

    return (
        <div className="stack">
            <section className="card">
                <div className="section-head">
                    <h2>Мой прогресс</h2>
                    <span className="pill">{logs.length}</span>
                </div>

                <div className="stats-card">
                    <div className="stat-item">
                        <div className="stat-label">Текущий вес</div>
                        <div className="stat-value">{currentWeight || '—'}</div>
                        <div className="stat-hint">кг</div>
                    </div>

                    <div className="stat-item">
                        <div className="stat-label">Начальный вес</div>
                        <div className="stat-value">{startWeight || '—'}</div>
                        <div className="stat-hint">кг</div>
                    </div>

                    <div className="stat-item">
                        <div className="stat-label">Цель</div>
                        <div className="stat-value">{targetWeight || '—'}</div>
                        <div className="stat-hint">кг</div>
                    </div>

                    <div className="stat-item">
                        <div className="stat-label">Изменение</div>
                        <div className="stat-value">
                            {change > 0 ? `-${change.toFixed(1)}` : change < 0 ? `+${Math.abs(change).toFixed(1)}` : '0'}
                        </div>
                        <div className="stat-hint">кг</div>
                    </div>
                </div>
            </section>

            <section className="card">
                <div className="section-head">
                    <h2>Записать вес</h2>
                </div>

                <div className="form-grid">
                    <label className="form-field">
                        <span>Ваш вес сегодня, кг</span>
                        <input
                            type="number"
                            value={weight}
                            placeholder="Например: 55.5"
                            onChange={(e) => setWeight(e.target.value)}
                        />
                    </label>
                </div>

                <button className="primary-btn full-btn" onClick={saveWeight} disabled={saving}>
                    {saving ? 'Сохраняем...' : 'Сохранить вес'}
                </button>
            </section>

            <section className="card">
                <div className="section-head">
                    <h2>История веса</h2>
                </div>

                {logs.length === 0 ? (
                    <p className="muted">Пока нет записей. Добавьте первый вес выше.</p>
                ) : (
                    <div className="weight-list">
                        {[...logs].reverse().map((log, index) => {
                            const value = log.Weight ?? log.weight ?? 0
                            const date = log.Date ?? log.date ?? log.CreatedAt ?? log.createdAt ?? ''

                            return (
                                <div className="weight-row" key={log.ID ?? log.id ?? index}>
                                    <div>
                                        <strong>{value} кг</strong>
                                        <p>{date ? new Date(date).toLocaleDateString('ru-RU') : 'Дата не указана'}</p>
                                    </div>
                                    <span>⚖️</span>
                                </div>
                            )
                        })}
                    </div>
                )}
            </section>
        </div>
    )
}