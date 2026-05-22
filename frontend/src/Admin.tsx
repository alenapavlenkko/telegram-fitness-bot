import * as React from 'react'

// ADMIN PANEL
export default function Admin() {

    // Название тренировки
    const [title, setTitle] =
        React.useState('')

    // Описание тренировки
    const [description, setDescription] =
        React.useState('')

    // Длительность тренировки
    const [duration, setDuration] =
        React.useState('')

    // Сложность
    const [difficulty, setDifficulty] =
        React.useState('')

    // YouTube ссылка
    const [youtubeLink, setYoutubeLink] =
        React.useState('')

    // Состояние загрузки
    const [loading, setLoading] =
        React.useState(false)

    // Сообщение об ошибке
    const [error, setError] =
        React.useState<string | null>(null)

    // Сообщение об успехе
    const [success, setSuccess] =
        React.useState<string | null>(null)

    async function createTraining() {

        // Сброс сообщений
        setError(null)
        setSuccess(null)

        // Проверка обязательных полей
        if (
            !title ||
            !description ||
            !duration
        ) {

            setError(
                'Заполните все обязательные поля',
            )

            return
        }

        try {

            // Включаем loading
            setLoading(true)

            // Отправляем POST запрос
            const response = await fetch(

                '/api/admin/trainings',

                {
                    method: 'POST',

                    headers: {
                        'Content-Type': 'application/json',
                    },

                    body: JSON.stringify({

                        title,

                        description,

                        duration: Number(duration),

                        difficulty,

                        youtube_link: youtubeLink,
                    }),
                },
            )

            // Если ошибка API
            if (!response.ok) {

                throw new Error(
                    'Ошибка при создании тренировки',
                )
            }

            setSuccess(
                '✅ Тренировка успешно добавлена',
            )

            // Очистка формы
            setTitle('')
            setDescription('')
            setDuration('')
            setDifficulty('')
            setYoutubeLink('')
        }
        catch (err: any) {

            setError(
                err.message,
            )
        }
        finally {

            // Выключаем loading
            setLoading(false)
        }
    }

    // RENDER
    return (

        <div className="stack">

            <section className="card">

                {/* HEADER */}

                <div className="section-head">

                    <h2>
                        👑 Admin Panel
                    </h2>

                </div>

                <p className="muted">

                    Управление тренировками платформы

                </p>

                {/* SUCCESS */}

                {success && (

                    <div className="success-box">
                        {success}
                    </div>
                )}

                {/* ERROR */}

                {error && (

                    <div className="error-box">
                        {error}
                    </div>
                )}

                {/* FORM */}

                <div className="form-group">

                    {/* TITLE */}

                    <input
                        className="input"
                        placeholder="Название тренировки"
                        value={title}
                        onChange={(e) =>
                            setTitle(
                                e.target.value,
                            )
                        }
                    />

                    {/* DESCRIPTION */}

                    <textarea
                        className="input"
                        placeholder="Описание тренировки"
                        value={description}
                        onChange={(e) =>
                            setDescription(
                                e.target.value,
                            )
                        }
                    />

                    {/* DURATION */}

                    <input
                        className="input"
                        placeholder="Длительность (мин)"
                        value={duration}
                        onChange={(e) =>
                            setDuration(
                                e.target.value,
                            )
                        }
                    />

                    {/* DIFFICULTY */}

                    <input
                        className="input"
                        placeholder="Сложность"
                        value={difficulty}
                        onChange={(e) =>
                            setDifficulty(
                                e.target.value,
                            )
                        }
                    />

                    {/* YOUTUBE */}

                    <input
                        className="input"
                        placeholder="YouTube ссылка"
                        value={youtubeLink}
                        onChange={(e) =>
                            setYoutubeLink(
                                e.target.value,
                            )
                        }
                    />

                    {/* BUTTON */}

                    <button
                        className="primary-btn"
                        onClick={createTraining}
                        disabled={loading}
                    >

                        {loading
                            ? '⏳ Сохранение...'
                            : '➕ Добавить тренировку'}

                    </button>

                </div>

            </section>

        </div>
    )
}