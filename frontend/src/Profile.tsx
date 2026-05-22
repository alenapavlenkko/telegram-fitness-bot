import * as React from 'react'

type ProfileForm = {
    telegramId: number
    name: string
    age: number
    gender: string
    height: number
    weight: number
    goal: string
    activity: string
    fitnessLevel: string
    targetWeight: number
}

const defaultTelegramId = 898030333

function normalizeUser(user: any, fallback: ProfileForm): ProfileForm {
    return {
        telegramId: user.TelegramID ?? user.telegramId ?? fallback.telegramId,
        name: user.Name ?? user.name ?? fallback.name,
        age: user.Age ?? user.age ?? fallback.age,
        gender: user.Gender ?? user.gender ?? fallback.gender,
        height: user.Height ?? user.height ?? fallback.height,
        weight: user.Weight ?? user.weight ?? fallback.weight,
        goal: user.Goal ?? user.goal ?? fallback.goal,
        activity: user.Activity ?? user.activity ?? fallback.activity,
        fitnessLevel: user.FitnessLevel ?? user.fitnessLevel ?? fallback.fitnessLevel,
        targetWeight: user.TargetWeight ?? user.targetWeight ?? fallback.targetWeight,
    }
}

export default function Profile({
                                    setPage,
                                }: {
    setPage: (page: any) => void
}) {
    const tgUser = (window as any).Telegram?.WebApp?.initDataUnsafe?.user

    const initialForm: ProfileForm = {
        telegramId: tgUser?.id ?? defaultTelegramId,
        name: tgUser?.first_name ?? 'Алёна',
        age: 20,
        gender: 'female',
        height: 165,
        weight: 55,
        goal: 'Похудение',
        activity: 'Средняя',
        fitnessLevel: 'Начальный',
        targetWeight: 52,
    }

    const [form, setForm] = React.useState<ProfileForm>(initialForm)
    const [loading, setLoading] = React.useState(true)
    const [saving, setSaving] = React.useState(false)
    const [saved, setSaved] = React.useState(false)

    React.useEffect(() => {
        async function loadProfile() {
            try {
                const res = await fetch(`/api/profile/${initialForm.telegramId}`)

                if (res.status === 404) {
                    return
                }

                if (!res.ok) {
                    throw new Error(`Ошибка загрузки: ${res.status}`)
                }

                const user = await res.json()
                setForm((prev) => normalizeUser(user, prev))
            } catch (e) {
                console.log(e)
            } finally {
                setLoading(false)
            }
        }

        loadProfile()
    }, [initialForm.telegramId])

    function updateField<K extends keyof ProfileForm>(key: K, value: ProfileForm[K]) {
        setSaved(false)
        setForm((prev) => ({
            ...prev,
            [key]: value,
        }))
    }

    async function saveProfile() {
        setSaving(true)
        setSaved(false)

        try {
            const res = await fetch('/api/profile', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(form),
            })

            if (!res.ok) {
                throw new Error(`Ошибка сохранения: ${res.status}`)
            }

            const updatedUser = await res.json()
            setForm((prev) => normalizeUser(updatedUser, prev))
            setSaved(true)
        } catch (e) {
            alert(e instanceof Error ? e.message : 'Ошибка сохранения профиля')
        } finally {
            setSaving(false)
        }
    }

    if (loading) {
        return (
            <div className="stack">
                <section className="card">
                    <h2>Профиль</h2>
                    <p className="muted">Загрузка профиля...</p>
                </section>
            </div>
        )
    }

    return (
        <div className="stack">
            <section className="card profile-card">
                <div className="avatar">
                    {form.name ? form.name.slice(0, 1).toUpperCase() : 'U'}
                </div>
                <div>
                    <h2 className="profile-name">{form.name || 'Профиль'}</h2>
                    <p className="profile-subtitle">
                        {form.age} лет · {form.height} см · {form.weight} кг
                    </p>
                </div>
            </section>

            <section className="card">
                <div className="section-head">
                    <h2>Редактировать профиль</h2>
                </div>

                <div className="form-grid">
                    <label className="form-field">
                        <span>Имя</span>
                        <input
                            value={form.name}
                            onChange={(e) => updateField('name', e.target.value)}
                            placeholder="Введите имя"
                        />
                    </label>

                    <label className="form-field">
                        <span>Возраст</span>
                        <input
                            type="number"
                            value={form.age}
                            onChange={(e) => updateField('age', Number(e.target.value))}
                        />
                    </label>

                    <label className="form-field">
                        <span>Пол</span>
                        <select
                            value={form.gender}
                            onChange={(e) => updateField('gender', e.target.value)}
                        >
                            <option value="female">Женский</option>
                            <option value="male">Мужской</option>
                            <option value="other">Другое</option>
                        </select>
                    </label>

                    <label className="form-field">
                        <span>Рост, см</span>
                        <input
                            type="number"
                            value={form.height}
                            onChange={(e) => updateField('height', Number(e.target.value))}
                        />
                    </label>

                    <label className="form-field">
                        <span>Вес, кг</span>
                        <input
                            type="number"
                            value={form.weight}
                            onChange={(e) => updateField('weight', Number(e.target.value))}
                        />
                    </label>

                    <label className="form-field">
                        <span>Желаемый вес, кг</span>
                        <input
                            type="number"
                            value={form.targetWeight}
                            onChange={(e) => updateField('targetWeight', Number(e.target.value))}
                        />
                    </label>

                    <label className="form-field">
                        <span>Цель</span>
                        <select
                            value={form.goal}
                            onChange={(e) => updateField('goal', e.target.value)}
                        >
                            <option>Похудение</option>
                            <option>Набор массы</option>
                            <option>Поддержание формы</option>
                            <option>Улучшение выносливости</option>
                        </select>
                    </label>

                    <label className="form-field">
                        <span>Активность</span>
                        <select
                            value={form.activity}
                            onChange={(e) => updateField('activity', e.target.value)}
                        >
                            <option>Низкая</option>
                            <option>Средняя</option>
                            <option>Высокая</option>
                        </select>
                    </label>

                    <label className="form-field">
                        <span>Уровень подготовки</span>
                        <select
                            value={form.fitnessLevel}
                            onChange={(e) => updateField('fitnessLevel', e.target.value)}
                        >
                            <option>Начальный</option>
                            <option>Средний</option>
                            <option>Продвинутый</option>
                        </select>
                    </label>
                </div>

                {saved && <p className="success-text">✅ Профиль сохранён и обновлён</p>}

                <button className="primary-btn full-btn" onClick={saveProfile} disabled={saving}>
                    {saving ? 'Сохраняем...' : 'Сохранить профиль'}
                </button>
            </section>
        </div>
    )
}