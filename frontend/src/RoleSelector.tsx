import * as React from 'react'

type Props = {
    isAdmin: boolean
    onSelect: (
        role: 'user' | 'admin'
    ) => void
}

export default function RoleSelector({
                                         isAdmin,
                                         onSelect,
                                     }: Props) {

    return (

        <div className="stack">

            <section className="card">

                <div className="section-head">
                    <h2>Выберите режим</h2>
                </div>

                <div className="quick-grid">

                    {/* USER */}

                    <button
                        className="quick-card"
                        onClick={() => onSelect('user')}
                    >

                        <div className="quick-icon">
                            👤
                        </div>

                        <div className="quick-title">
                            Пользователь
                        </div>

                        <div className="quick-desc">
                            Тренировки, питание и статистика
                        </div>

                    </button>

                    {/* ADMIN */}

                    {isAdmin && (

                        <button
                            className="quick-card"
                            onClick={() => onSelect('admin')}
                        >

                            <div className="quick-icon">
                                👑
                            </div>

                            <div className="quick-title">
                                Администратор
                            </div>

                            <div className="quick-desc">
                                Управление платформой
                            </div>

                        </button>
                    )}

                </div>

            </section>

        </div>
    )
}